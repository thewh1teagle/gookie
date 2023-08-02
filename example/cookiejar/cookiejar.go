package main

import (
	"fmt"
	"net/http"

	"github.com/thewh1teagle/gookie/gookie"
	"golang.org/x/net/html"
)

func main() {
	cookies := gookie.Chrome()
	jar := gookie.ToCookieJar(cookies)

	client := &http.Client{
		Jar: jar, // try to comment out
	}
	loginURL := "https://github.com/settings/profile"
	req, err := http.NewRequest(http.MethodGet, loginURL, nil)
	if err != nil {
		fmt.Printf("Error creating request: %s\n", err)
		return
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error performing HTTP request: %s\n", err)
		return
	}
	defer res.Body.Close()

	// Check if the response status code is 200 (OK)
	if res.StatusCode != http.StatusOK {
		fmt.Printf("Request failed with status code: %d\n", res.StatusCode)
		return
	}

	// Parse the HTML content of the response body
	doc, err := html.Parse(res.Body)
	if err != nil {
		fmt.Printf("Error parsing HTML: %s\n", err)
		return
	}

	// Find the title element in the HTML document
	title := findTitle(doc)

	// Print the title
	if title != "" {
		fmt.Println("Title:", title)
	} else {
		fmt.Println("Title not found.")
	}
}

// Helper function to find the title in the HTML document
func findTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
		return n.FirstChild.Data
	}

	// Recursively search for the title element in the HTML tree
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if title := findTitle(c); title != "" {
			return title
		}
	}

	return ""
}
