package main

import (
	"net/url"
	"path/filepath"
)

func normalizeURL(parseURL string) (string, error) {
	parsedURL, err := url.Parse(parseURL)
	if err != nil {
		return "", err
	}

	parsedURLString, err :=
		url.JoinPath(parsedURL.Host, filepath.Clean(parsedURL.Path))

	if err != nil {
		return "", err
	}

	return parsedURLString, nil
}
