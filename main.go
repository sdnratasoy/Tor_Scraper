package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"github.com/chromedp/chromedp"
	"golang.org/x/net/proxy"
)

func readTargets(fileName string) ([]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var targets []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			targets = append(targets, line)
		}
	}
	return targets, nil
}

func createTorClient() (*http.Client, error) {
	ports := []string{"127.0.0.1:9150", "127.0.0.1:9050"}

	for _, port := range ports {
		dialer, err := proxy.SOCKS5("tcp", port, nil, proxy.Direct)
		if err == nil {
			return &http.Client{
				Transport: &http.Transport{Dial: dialer.Dial},
				Timeout:   30 * time.Second,
			}, nil
		}
	}
	return nil, fmt.Errorf("Tor bulunamadı")
}

func scanURL(client *http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func saveHTML(url string, content []byte, dir string) error {
	name := strings.NewReplacer(
		"http://", "", "https://", "", "/", "_", ".onion", "", ":", "_",
	).Replace(url)

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	path := fmt.Sprintf("%s/%s_%s.html", dir, name, timestamp)
	return os.WriteFile(path, content, 0644)
}

func takeScreenshot(url, dir string) error {
	name := strings.NewReplacer(
		"http://", "", "https://", "", "/", "_", ".onion", "", ":", "_",
	).Replace(url)

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	path := fmt.Sprintf("%s/%s_%s.png", dir, name, timestamp)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ProxyServer("socks5://127.0.0.1:9150"),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx,
		chromedp.WithErrorf(func(string, ...interface{}) {}),
		chromedp.WithLogf(func(string, ...interface{}) {}),
	)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(2*time.Second),
		chromedp.FullScreenshot(&buf, 100), // Tüm sayfanın screenshot'ını al
	)
	if err != nil {
		return err
	}
	return os.WriteFile(path, buf, 0644)
}

func writeLog(file *os.File, status, url string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(file, "[%s] %s %s\n", timestamp, status, url)
}

func main() {
	const exampleYAML = `# TorScraper Hedef Listesi
http://example.onion
`

	fmt.Println("==============================================")
	fmt.Println("      TOR SCRAPER")
	fmt.Println("==============================================\n")

	os.Mkdir("output", 0755)

	logFile, err := os.Create("scan_report.log")
	if err != nil {
		fmt.Println("[HATA] Log oluşturulamadı:", err)
		return
	}
	defer logFile.Close()
	writeLog(logFile, "BAŞLANGIÇ", "Tarama başladı")

	fmt.Println("[INFO] Tor proxy'sine bağlanılıyor (9150/9050)")
	client, err := createTorClient()
	if err != nil {
		fmt.Println("[HATA] Tor bulunamadı!")
		fmt.Println("Çözüm: Tor Browser'ı başlatın (https://www.torproject.org/download/)")
		return
	}
	fmt.Println("[BAŞARILI] Tor proxy'sine bağlandı!\n")

	fmt.Println("[INFO] targets.yaml okunuyor.")
	targets, err := readTargets("targets.yaml")
	if err != nil {
		fmt.Println("[HATA] targets.yaml bulunamadı, örnek oluşturuluyor.")
		os.WriteFile("targets.yaml", []byte(exampleYAML), 0644)
		fmt.Println("[BAŞARILI] targets.yaml oluşturuldu. Düzenleyip tekrar çalıştırın.")
		return
	}

	if len(targets) == 0 {
		fmt.Println("[UYARI] Hedef listesi boş!")
		return
	}

	fmt.Printf("[INFO] %d hedef bulundu.\n\n", len(targets))
	fmt.Println("==============================================")
	fmt.Println("      TARAMA BAŞLIYOR")
	fmt.Println("==============================================\n")

	success, fail := 0, 0

	for i, url := range targets {
		fmt.Printf("[%d/%d] Taranıyor: %s\n", i+1, len(targets), url)

		html, err := scanURL(client, url)
		if err != nil {
			if strings.Contains(err.Error(), "connectex") || strings.Contains(err.Error(), "connection refused") {
				fmt.Printf("  └─ [HATA] Tor bağlantısı kesildi!\n")
				fmt.Printf("           Tor Browser'ı kontrol edin ve tekrar başlatın.\n\n")
				writeLog(logFile, "FAIL-TOR", url)
			} else {
				fmt.Printf("  └─ [BAŞARISIZ] %s\n\n", err)
				writeLog(logFile, "FAIL", url)
			}
			fail++
			continue
		}
		if err := saveHTML(url, html, "output"); err != nil {
			fmt.Printf("  └─ [HATA] Kayıt başarısız\n\n")
			writeLog(logFile, "FAIL", url)
			fail++
			continue
		}

		fmt.Printf("  ├─ [BAŞARILI] HTML kaydedildi (%d bytes)\n", len(html))
		fmt.Printf("  └─ [BİLGİ] Screenshot alınıyor...\n")

		if err := takeScreenshot(url, "output"); err == nil {
			fmt.Printf("     └─ [BAŞARILI] Screenshot kaydedildi\n")
		}

		writeLog(logFile, "SUCCESS", url)
		success++
		fmt.Println()
		time.Sleep(1 * time.Second)
	}

	fmt.Println("==============================================")
	fmt.Println("       TARAMA TAMAMLANDI")
	fmt.Println("==============================================")
	fmt.Printf("Toplam:     %d\n", len(targets))
	fmt.Printf("Başarılı:   %d\n", success)
	fmt.Printf("Başarısız:  %d\n\n", fail)
	fmt.Println("Sonuçlar: output/ & scan_report.log")
	fmt.Println("==============================================")

	writeLog(logFile, "BİTİŞ", fmt.Sprintf("Başarılı: %d, Başarısız: %d", success, fail))
}
