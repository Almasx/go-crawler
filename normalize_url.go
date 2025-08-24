package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func normalizeURL(urlString string) (string, error) {
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}

	normalizedURL := "https://" + parsedURL.Host + parsedURL.Path
	return normalizedURL, nil
}

func getURLsFromHTML(htmlBody string, baseURL string) ([]string, error) {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		return nil, err
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	var urls []string
	var traverse func(n *html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					hrefURL, err := url.Parse(attr.Val)
					if err == nil {
						resolvedURL := base.ResolveReference(hrefURL)
						urls = append(urls, resolvedURL.String())
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(doc)

	return urls, nil
}

func getHTML(rawURL string) (string, error) {
	resp, err := http.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "text/html") {
		return "", fmt.Errorf("content type is not text/html: %s", contentType)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (cfg *config) crawlPage(rawCurrentURL string) {
	cfg.concurrencyControl <- struct{}{}
	defer func() {
		<-cfg.concurrencyControl
		cfg.wg.Done()
	}()

	cfg.mu.Lock()
	if len(cfg.pages) >= cfg.maxPages {
		cfg.mu.Unlock()
		return
	}
	cfg.mu.Unlock()

	normalizedCurrentURL, err := normalizeURL(rawCurrentURL)
	if err != nil {
		fmt.Println("Error normalizing current URL:", err)
		return
	}

	if !cfg.addPageVisit(normalizedCurrentURL) {
		return
	}

	// Check if URL is on same domain
	currentURL, err := url.Parse(rawCurrentURL)
	if err != nil {
		fmt.Println("Error parsing current URL:", err)
		return
	}

	if currentURL.Host != cfg.baseURL.Host {
		return
	}

	fmt.Printf("crawling %s\n", normalizedCurrentURL)

	htmlBody, err := getHTML(rawCurrentURL)
	if err != nil {
		fmt.Println("Error getting HTML:", err)
		return
	}

	urls, err := getURLsFromHTML(htmlBody, rawCurrentURL)

	if err != nil {
		fmt.Println("Error getting URLs from HTML:", err)
		return
	}

	for _, linkURL := range urls {
		// Check if this is an internal link before adding to count
		parsedURL, err := url.Parse(linkURL)
		if err != nil {
			continue
		}
		
		// Only count internal links (same domain)
		if parsedURL.Host == cfg.baseURL.Host {
			normalizedURL, err := normalizeURL(linkURL)
			if err == nil {
				cfg.addPageLink(normalizedURL)
			}
		}
		
		cfg.wg.Add(1)
		go cfg.crawlPage(linkURL)
	}
}

func (cfg *config) addPageVisit(normalizedURL string) (isFirst bool) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if cfg.pages[normalizedURL] > 0 {
		return false
	}

	cfg.pages[normalizedURL] = 1
	return true
}

func (cfg *config) addPageLink(normalizedURL string) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	
	cfg.pages[normalizedURL]++
}
