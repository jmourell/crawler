package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type config struct {
	pages              map[string]int
	maxPages           int
	baseURL            *url.URL
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
}

func (cfg *config) addPageVisit(normalizedURL string) (isFirst bool) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	_, ok := cfg.pages[normalizedURL]
	if ok {
		cfg.pages[normalizedURL]++
		return false
	}
	cfg.pages[normalizedURL] = 1
	return true
}
func (cfg *config) checkMaxPages() (maxPagesReached bool) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	return cfg.maxPages <= len(cfg.pages)
}

func (cfg *config) crawlPage(rawCurrentURL string) {
	cfg.concurrencyControl <- struct{}{}
	defer func() {
		<-cfg.concurrencyControl
		cfg.wg.Done()
	}()
	if cfg.checkMaxPages() {
		return
	}

	currentURL, err := url.Parse(rawCurrentURL)
	if err != nil {
		fmt.Println("Error found in URL")
		return
	}
	if currentURL.Host != cfg.baseURL.Host {
		return
	}
	currentURLString, err := normalizeURL(rawCurrentURL)
	if err != nil {
		fmt.Println("Normalizing failed")
		return
	}
	isNew := cfg.addPageVisit(currentURLString)

	if !isNew {
		return
	}

	if cfg.checkMaxPages() {
		return
	}

	fmt.Printf("crawling %s\n", rawCurrentURL)

	html, err := getHTML(rawCurrentURL)
	if err != nil {
		fmt.Printf("error getting html: %v", err)
		return
	}

	urls, err := getURLsFromHTML(html, cfg.baseURL.String())
	if err != nil {
		fmt.Println("error parsing urls")
		return
	}
	for _, address := range urls {
		if cfg.checkMaxPages() {
			return
		}
		cfg.wg.Add(1)
		go cfg.crawlPage(address)
	}

}

func getHTML(rawURL string) (string, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 400 {
		return "", fmt.Errorf("error in status: %s", resp.Status)
	}
	if !strings.Contains(strings.ToLower(resp.Header.Get("content-type")), "text/html") {
		return "", fmt.Errorf("non text/html content type : %s", strings.ToLower(resp.Header.Get("content-type")))
	}
	data, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func printReport(pages map[string]int, baseURL string) {
	fmt.Println("=============================")
	fmt.Printf("REPORT for %s", baseURL)
	fmt.Println("=============================")

	keys := make([]string, 0, len(pages))

	for key := range pages {
		keys = append(keys, key)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		if pages[keys[i]] == pages[keys[i]] {
			return strings.Compare(keys[i], keys[j]) == -1
		}
		return pages[keys[i]] < pages[keys[j]]
	})

	for _, v := range keys {
		fmt.Printf("Found %d internal links to %s\n", pages[v], v)
	}
}

func main() {
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) < 3 {
		fmt.Println("no website, max concurrency or max pages provided")
		os.Exit(1)
	}
	if len(argsWithoutProg) > 3 {
		fmt.Println("too many arguments provided")
		os.Exit(1)
	}
	fmt.Printf("starting crawl of: %s\n", argsWithoutProg[0])
	pages := make(map[string]int)
	baseURL, err := url.Parse(argsWithoutProg[0])
	if err != nil {
		fmt.Println("Error found in URL")
		return
	}
	channels, err := strconv.Atoi(argsWithoutProg[1])
	if err != nil {
		fmt.Printf("error found in Max concurrency: %s", argsWithoutProg[1])
		return
	}
	maxPages, err := strconv.Atoi(argsWithoutProg[2])
	if err != nil {
		fmt.Printf("error found in Max pages: %s", argsWithoutProg[2])
		return
	}
	cfg := config{
		pages:              pages,
		baseURL:            baseURL,
		mu:                 &sync.Mutex{},
		maxPages:           maxPages,
		concurrencyControl: make(chan struct{}, channels),
		wg:                 &sync.WaitGroup{},
	}
	fmt.Printf("Max Pages: %d Max Concurrency: %d\n", maxPages, channels)
	cfg.wg.Add(1)
	cfg.crawlPage(argsWithoutProg[0])

	cfg.wg.Wait()
	printReport(cfg.pages, cfg.baseURL.String())
}
