package main

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func getURLsFromHTML(htmlBody, rawBaseURL string) ([]string, error) {
	urls := []string{}
	r := strings.NewReader(htmlBody)

	doc, err := html.Parse(r)
	if err != nil {
		return []string{}, err
	}
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					parsedURL, err := url.Parse(a.Val)
					if err != nil {
						continue
					}
					parsedURLString := parsedURL.String()
					if !parsedURL.IsAbs() {
						parsedURLString, err = url.JoinPath(rawBaseURL, parsedURL.Path)
						if err != nil {
							continue
						}
					}
					urls = append(urls, parsedURLString)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return urls, nil
}
