package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type TelnetClient struct {
	host       string
	port       string
	timeout    time.Duration
	conn       net.Conn
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	shutdownMu sync.Mutex
	isShutdown bool
}

func NewTelnetClient(host, port string, timeout time.Duration) *TelnetClient {
	ctx, cancel := context.WithCancel(context.Background())
	return &TelnetClient{
		host:    host,
		port:    port,
		timeout: timeout,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (tc *TelnetClient) Connect() error {
	address := net.JoinHostPort(tc.host, tc.port)
	fmt.Fprintf(os.Stderr, "Подключение к %s...\n", address)

	conn, err := net.DialTimeout("tcp", address, tc.timeout)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к %s: %w", address, err)
	}

	tc.conn = conn
	fmt.Fprintf(os.Stderr, "Соединение установлено с %s\n", address)
	fmt.Fprintf(os.Stderr, "Для завершения нажмите Ctrl+D или Ctrl+C\n")
	fmt.Fprintf(os.Stderr, "---\n")

	return nil
}

func (tc *TelnetClient) Run() error {
	if tc.conn == nil {
		return fmt.Errorf("соединение не установлено")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 2)

	tc.wg.Add(1)
	go tc.readFromSocket(errChan)

	tc.wg.Add(1)
	go tc.writeToSocket(errChan)

	select {
	case <-sigChan:
		fmt.Fprintf(os.Stderr, "\nПолучен сигнал прерывания. Закрытие соединения...\n")
		tc.shutdown()

	case err := <-errChan:
		if err != nil && err != io.EOF {
			fmt.Fprintf(os.Stderr, "\nОшибка: %v\n", err)
		}
		tc.shutdown()

	case <-tc.ctx.Done():
		tc.shutdown()
	}

	tc.wg.Wait()

	return nil
}

func (tc *TelnetClient) readFromSocket(errChan chan<- error) {
	defer tc.wg.Done()

	reader := bufio.NewReader(tc.conn)
	buf := make([]byte, 4096)

	for {
		select {
		case <-tc.ctx.Done():
			return
		default:
		}

		tc.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

		n, err := reader.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}

			if err == io.EOF {
				fmt.Fprintf(os.Stderr, "\nСоединение закрыто сервером\n")
			}

			if !tc.isShutdownActive() {
				select {
				case errChan <- err:
				default:
				}
			}
			return
		}

		if n > 0 {
			if _, err := os.Stdout.Write(buf[:n]); err != nil {
				select {
				case errChan <- fmt.Errorf("ошибка записи в STDOUT: %w", err):
				default:
				}
				return
			}
		}
	}
}

func (tc *TelnetClient) writeToSocket(errChan chan<- error) {
	defer tc.wg.Done()

	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-tc.ctx.Done():
			return
		default:
		}

		input, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Fprintf(os.Stderr, "\nEOF (Ctrl+D) - закрытие соединения\n")
				select {
				case errChan <- io.EOF:
				default:
				}
				return
			}

			if !tc.isShutdownActive() {
				select {
				case errChan <- fmt.Errorf("ошибка чтения из STDIN: %w", err):
				default:
				}
			}
			return
		}

		if _, err := tc.conn.Write(input); err != nil {
			if !tc.isShutdownActive() {
				select {
				case errChan <- fmt.Errorf("ошибка записи в сокет: %w", err):
				default:
				}
			}
			return
		}
	}
}

func (tc *TelnetClient) shutdown() {
	tc.shutdownMu.Lock()
	defer tc.shutdownMu.Unlock()

	if tc.isShutdown {
		return
	}

	tc.isShutdown = true

	tc.cancel()

	if tc.conn != nil {
		tc.conn.Close()
	}
}

func (tc *TelnetClient) isShutdownActive() bool {
	tc.shutdownMu.Lock()
	defer tc.shutdownMu.Unlock()
	return tc.isShutdown
}

func (tc *TelnetClient) Close() error {
	tc.shutdown()
	tc.wg.Wait()
	return nil
}

func main() {
	host := flag.String("host", "", "Хост для подключения (обязательный)")
	port := flag.String("port", "", "Порт для подключения (обязательный)")
	timeout := flag.Duration("timeout", 10*time.Second, "Таймаут подключения")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Использование: %s [опции]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Примитивный telnet-клиент для подключения к TCP-серверам\n\n")
		fmt.Fprintf(os.Stderr, "Опции:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nПримеры:\n")
		fmt.Fprintf(os.Stderr, "  %s -host=telehack.com -port=23\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -host=localhost -port=8080 -timeout=5s\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -host=smtp.gmail.com -port=25\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nДля завершения нажмите Ctrl+D или Ctrl+C\n")
	}

	flag.Parse()

	if *host == "" || *port == "" {
		fmt.Fprintf(os.Stderr, "Ошибка: необходимо указать host и port\n\n")
		flag.Usage()
		os.Exit(1)
	}

	client := NewTelnetClient(*host, *port, *timeout)
	defer client.Close()

	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка подключения: %v\n", err)
		os.Exit(1)
	}

	if err := client.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка выполнения: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Соединение закрыто\n")
}