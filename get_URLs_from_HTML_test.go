package main

import (
	"reflect"
	"testing"
)

func TestGetURLsFromHTML(t *testing.T) {
	tests := []struct {
		name      string
		inputURL  string
		inputBody string
		expected  []string
	}{
		{
			name:     "absolute and relative URLs",
			inputURL: "https://blog.boot.dev",
			inputBody: `
		<html>
			<body>
				<a href="/path/one">
					<span>Boot.dev</span>
				</a>
				<a href="https://other.com/path/one">
					<span>Boot.dev</span>
				</a>
			</body>
		</html>
		`,
			expected: []string{"https://blog.boot.dev/path/one", "https://other.com/path/one"},
		},
		{
			name:     "test nested URLs",
			inputURL: "https://blog.boot.dev",
			inputBody: `
		<html>
			<body>

				<span> </span>
				<div>
					<a href="/path/one">
						<span>Boot.dev</span>
					</a>
				</div>
				<span> </span>
				<a href="https://other.com/path/one">
					<span>Boot.dev</span>
				</a>
			</body>
		</html>
		`,
			expected: []string{"https://blog.boot.dev/path/one", "https://other.com/path/one"},
		},
		{
			name:     "test no URLs",
			inputURL: "https://blog.boot.dev",
			inputBody: `
		<html>
			<body>
				<span> </span>
				<div>
				</div>
				<span> </span>
			</body>
		</html>
		`,
			expected: []string{},
		},
		{
			name:     "test only URLs",
			inputURL: "https://blog.boot.dev",
			inputBody: `

				<a href="/path/one">
						<span>Boot.dev</span>
					</a>
				<a href="https://other.com/path/one">
					<span>Boot.dev</span>
				</a>
		`,
			expected: []string{"https://blog.boot.dev/path/one", "https://other.com/path/one"},
		},
	}

	for i, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := getURLsFromHTML(tc.inputBody, tc.inputURL)
			if err != nil {
				t.Errorf("Test %v - '%s' FAIL: unexpected error: %v", i, tc.name, err)
				return
			}
			if !reflect.DeepEqual(tc.expected, actual) {
				t.Errorf("Test %v - %s FAIL: expected URL: %v, actual: %v", i, tc.name, tc.expected, actual)
			}

		})
	}
}
