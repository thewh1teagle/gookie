package gookie

import (
	"fmt"
	"time"

	"net/http"
	"net/http/cookiejar"
	"net/url"

	_ "github.com/mattn/go-sqlite3"
)

type Cookie struct {
	Host       string    `json:"hostKey"`
	Path       string    `json:"path"`
	IsSecure   bool      `json:"isSecure"`
	Expires    time.Time `json:"Expires"`
	Name       string    `json:"name"`
	Value      string    `json:"value"`
	IsHttpOnly bool      `json:"isHttpOnly"`
	SameSite   int       `json:"sameSite"`
}

func ToCookieJar(cookies []Cookie) http.CookieJar {
	jar, err := cookiejar.New(nil)
	checkError(err)

	cookiesMap := make(map[string][]*http.Cookie)
	for _, cookie := range cookies {
		// Convert RawChromeCookie to http.Cookie
		httpCookie := &http.Cookie{
			Name:    cookie.Name,
			Value:   cookie.Value,
			Expires: cookie.Expires,
		}

		// Add the cookie to the cookiesMap under its HostKey

		cookiesMap[cookie.Host] = append(cookiesMap[cookie.Host], httpCookie)
	}

	// Set the cookies in the jar for each host
	for host, hostCookies := range cookiesMap {
		u, err := url.Parse(fmt.Sprintf("http://%v", host))
		if err != nil {
			fmt.Printf("Error parsing URL %s: %s\n", host, err)
			continue
		}
		fmt.Printf("set cookies %v %v", u, hostCookies)
		fmt.Println(u)
		jar.SetCookies(u, hostCookies)
	}
	return jar
}
