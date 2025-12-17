package main

import (
	"flag"
	"fmt"
	"os"
	"time"
	"wget-go/config"
	"wget-go/downloader"
)

func main() {
	// Параметры командной строки
	targetURL := flag.String("url", "", "URL для загрузки (обязательный)")
	depth := flag.Int("depth", 2, "Глубина рекурсии")
	concurrent := flag.Int("concurrent", 5, "Количество одновременных загрузок")
	timeout := flag.Duration("timeout", 30*time.Second, "Таймаут запроса")
	outputDir := flag.String("output", "downloaded_site", "Выходная директория")
	userAgent := flag.String("user-agent", "SimplifiedWget/1.0", "User-Agent")
	respectRobots := flag.Bool("respect-robots", true, "Соблюдать robots.txt")

	flag.Parse()

	if *targetURL == "" {
		fmt.Println("Использование:")
		fmt.Println("  go run main.go -url=https://example.com [опции]")
		fmt.Println("\nОпции:")
		flag.PrintDefaults()
		fmt.Println("\nПример:")
		fmt.Println("  go run main.go -url=https://example.com -depth=3 -respect-robots=true")
		os.Exit(1)
	}

	config := config.Config{
		MaxDepth:         *depth,
		MaxConcurrent:    *concurrent,
		Timeout:          *timeout,
		OutputDir:        *outputDir,
		UserAgent:        *userAgent,
		RespectRobotsTxt: *respectRobots,
	}

	downloader, err := downloader.NewDownloader(config, *targetURL)
	if err != nil {
		fmt.Printf("Ошибка инициализации: %v\n", err)
		os.Exit(1)
	}

	if err := downloader.Download(); err != nil {
		fmt.Printf("Ошибка загрузки: %v\n", err)
		os.Exit(1)
	}
}