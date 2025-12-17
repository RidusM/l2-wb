package downloader

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"wget-go/config"
	robots "wget-go/robots_parser"

	"golang.org/x/net/html"
)

type Downloader struct {
	config        config.Config
	baseURL       *url.URL
	visited       map[string]bool
	mu            sync.Mutex
	semaphore     chan struct{}
	wg            sync.WaitGroup
	client        *http.Client
	robotsParser  *robots.RobotsParser
	lastRequest   time.Time
	requestMu     sync.Mutex
}

// NewDownloader создает новый экземпляр загрузчика
func NewDownloader(cfg config.Config, baseURL string) (*Downloader, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	d := &Downloader{
		config:    cfg,
		baseURL:   u,
		visited:   make(map[string]bool),
		semaphore: make(chan struct{}, cfg.MaxConcurrent),
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
	}

	// Загружаем и парсим robots.txt если нужно
	if cfg.RespectRobotsTxt {
		if err := d.loadRobotsTxt(); err != nil {
			fmt.Printf("Предупреждение: не удалось загрузить robots.txt: %v\n", err)
		}
	}

	return d, nil
}

// loadRobotsTxt загружает и парсит robots.txt
func (d *Downloader) loadRobotsTxt() error {
	robotsURL := fmt.Sprintf("%s://%s/robots.txt", d.baseURL.Scheme, d.baseURL.Host)
	fmt.Printf("Загрузка robots.txt: %s\n", robotsURL)

	req, err := http.NewRequest("GET", robotsURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", d.config.UserAgent)

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		fmt.Println("robots.txt не найден - все пути разрешены")
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("статус: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	d.robotsParser = robots.NewRobotsParser(string(body))

	// Выводим информацию о правилах
	if d.robotsParser.DefaultRule != nil {
		fmt.Printf("Найдены правила robots.txt:\n")
		fmt.Printf("  Запрещенных путей: %d\n", len(d.robotsParser.DefaultRule.Disallow))
		if d.robotsParser.DefaultRule.CrawlDelay > 0 {
			fmt.Printf("  Задержка между запросами: %v\n", d.robotsParser.DefaultRule.CrawlDelay)
		}
	}

	return nil
}

// Download начинает процесс загрузки
func (d *Downloader) Download() error {
	fmt.Printf("Начало загрузки: %s\n", d.baseURL.String())
	fmt.Printf("Выходная директория: %s\n", d.config.OutputDir)
	fmt.Printf("Максимальная глубина: %d\n", d.config.MaxDepth)
	fmt.Printf("Соблюдение robots.txt: %v\n", d.config.RespectRobotsTxt)

	if err := os.MkdirAll(d.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("не удалось создать выходную директорию: %w", err)
	}

	d.downloadURL(d.baseURL.String(), 0)
	d.wg.Wait()

	fmt.Println("\nЗагрузка завершена!")
	return nil
}

// respectCrawlDelay соблюдает задержку между запросами
func (d *Downloader) respectCrawlDelay() {
	if d.robotsParser == nil {
		return
	}

	delay := d.robotsParser.GetCrawlDelay(d.config.UserAgent)
	if delay == 0 {
		return
	}

	d.requestMu.Lock()
	defer d.requestMu.Unlock()

	elapsed := time.Since(d.lastRequest)
	if elapsed < delay {
		time.Sleep(delay - elapsed)
	}
	d.lastRequest = time.Now()
}

// isAllowedByRobots проверяет, разрешен ли URL по robots.txt
func (d *Downloader) isAllowedByRobots(urlPath string) bool {
	if !d.config.RespectRobotsTxt || d.robotsParser == nil {
		return true
	}

	return d.robotsParser.IsAllowed(urlPath, d.config.UserAgent)
}

// downloadURL загружает URL и обрабатывает его содержимое
func (d *Downloader) downloadURL(urlStr string, depth int) {
	if depth > d.config.MaxDepth {
		return
	}

	d.mu.Lock()
	if d.visited[urlStr] {
		d.mu.Unlock()
		return
	}
	d.visited[urlStr] = true
	d.mu.Unlock()

	d.wg.Add(1)
	d.semaphore <- struct{}{}

	go func() {
		defer d.wg.Done()
		defer func() { <-d.semaphore }()

		if err := d.processURL(urlStr, depth); err != nil {
			fmt.Printf("Ошибка при загрузке %s: %v\n", urlStr, err)
		}
	}()
}

// processURL обрабатывает один URL
func (d *Downloader) processURL(urlStr string, depth int) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	// Проверяем, что URL принадлежит тому же домену
	if parsedURL.Host != "" && parsedURL.Host != d.baseURL.Host {
		return nil
	}

	// Если URL относительный, делаем его абсолютным
	if parsedURL.Host == "" {
		parsedURL = d.baseURL.ResolveReference(parsedURL)
		urlStr = parsedURL.String()
	}

	// Проверяем robots.txt
	if !d.isAllowedByRobots(parsedURL.Path) {
		fmt.Printf("[Глубина %d] Запрещено robots.txt: %s\n", depth, urlStr)
		return nil
	}

	fmt.Printf("[Глубина %d] Загрузка: %s\n", depth, urlStr)

	// Соблюдаем задержку
	d.respectCrawlDelay()

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", d.config.UserAgent)

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("статус: %d", resp.StatusCode)
	}

	// Читаем содержимое
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Сохраняем файл
	localPath := d.getLocalPath(parsedURL)
	if err := d.saveFile(localPath, body); err != nil {
		return err
	}

	// Если это HTML, парсим и извлекаем ссылки
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		links := d.extractLinks(body, parsedURL)
		for _, link := range links {
			d.downloadURL(link, depth+1)
		}
	}

	return nil
}

// extractLinks извлекает все ссылки из HTML
func (d *Downloader) extractLinks(htmlContent []byte, baseURL *url.URL) []string {
	var links []string
	seen := make(map[string]bool)

	doc, err := html.Parse(strings.NewReader(string(htmlContent)))
	if err != nil {
		return links
	}

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			var linkAttr string
			switch n.Data {
			case "a", "link":
				linkAttr = "href"
			case "script", "img":
				linkAttr = "src"
			case "source":
				linkAttr = "srcset"
			}

			if linkAttr != "" {
				for _, attr := range n.Attr {
					if attr.Key == linkAttr {
						link := d.normalizeURL(attr.Val, baseURL)
						if link != "" && !seen[link] {
							seen[link] = true
							links = append(links, link)
						}
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return links
}

// normalizeURL нормализует URL (преобразует относительные в абсолютные)
func (d *Downloader) normalizeURL(href string, baseURL *url.URL) string {
	href = strings.TrimSpace(href)
	if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "javascript:") ||
		strings.HasPrefix(href, "mailto:") || strings.HasPrefix(href, "tel:") {
		return ""
	}

	parsedURL, err := url.Parse(href)
	if err != nil {
		return ""
	}

	// Делаем URL абсолютным
	absoluteURL := baseURL.ResolveReference(parsedURL)

	// Проверяем, что это тот же домен
	if absoluteURL.Host != d.baseURL.Host {
		return ""
	}

	// Убираем фрагмент
	absoluteURL.Fragment = ""

	return absoluteURL.String()
}

// getLocalPath преобразует URL в локальный путь файла
func (d *Downloader) getLocalPath(u *url.URL) string {
	// Используем путь URL как базу
	localPath := u.Path
	if localPath == "" || localPath == "/" {
		localPath = "/index.html"
	}

	// Если путь заканчивается на /, добавляем index.html
	if strings.HasSuffix(localPath, "/") {
		localPath += "index.html"
	}

	// Если нет расширения, добавляем .html
	if path.Ext(localPath) == "" {
		localPath += ".html"
	}

	// Удаляем ведущий слэш и объединяем с выходной директорией
	localPath = strings.TrimPrefix(localPath, "/")
	return filepath.Join(d.config.OutputDir, localPath)
}

// saveFile сохраняет содержимое в файл
func (d *Downloader) saveFile(filePath string, content []byte) error {
	// Создаем директории
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Сохраняем файл
	return os.WriteFile(filePath, content, 0644)
}