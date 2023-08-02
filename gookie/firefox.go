package gookie

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"gopkg.in/ini.v1"
)

func queryCookiesFirefox(conn *sql.DB, optionalParams ...string) []Cookie {
	var domain string = "" // default value
	if len(optionalParams) > 0 {
		domain = optionalParams[0]
	}
	query := `
	SELECT host, path, isSecure, expiry, name, value, isHttpOnly, sameSite
	FROM moz_cookies 
	WHERE host like ?;
	`

	rows, err := conn.Query(query, "%"+domain+"%")
	checkError(err)
	defer rows.Close()
	defer conn.Close()

	var cookies []Cookie

	for rows.Next() {
		var host, path, name, value string
		var sameSite int
		var expires int64
		var isSecure, isHttpOnly bool

		err := rows.Scan(&host, &path, &isSecure, &expires, &name, &value, &isHttpOnly, &sameSite)
		if err != nil {
			fmt.Println("Error scanning row:", err)
			continue
		}
		cookie := Cookie{
			Host:     host,
			Path:     path,
			Secure:   isSecure,
			Expires:  time.Unix(expires, 0), // number of seconds since 1970
			Name:     name,
			Value:    value,
			HttpOnly: isHttpOnly,
			SameSite: sameSite,
		}
		cookies = append(cookies, cookie)
	}
	return cookies
}

func findFirefoxDataPath() (string, error) {
	switch runtime.GOOS {
	case "windows":
		appDataPath := os.Getenv("APPDATA")
		dataPath := filepath.Join(appDataPath, "/Mozilla/Firefox")
		return dataPath, nil
	case "linux":
		homePath, _ := os.UserHomeDir()
		linuxPaths := []string{
			filepath.Join(homePath, "/snap/firefox/common/.mozilla/firefox"),
			filepath.Join(homePath, "/.mozilla/firefox"),
		}
		for _, path := range linuxPaths {
			if pathExists(path) {
				return path, nil
			} else {
				fmt.Printf("path not exists %v\n", path)
			}
		}
	case "darwin":
		homePath, _ := os.UserHomeDir()
		path := filepath.Join(homePath, "/Library/Application Support/Firefox")
		if pathExists(path) {
			return path, nil
		}
	}
	return "", errors.New("User data not found for firefox browser")
}

func Firefox(params ...string) []Cookie {
	firefoxDataPath, _ := findFirefoxDataPath()
	fmt.Printf("Firefox path: %v\n", firefoxDataPath)
	cfg, err := ini.Load(filepath.Join(firefoxDataPath, "profiles.ini"))
	checkError(err)
	defaultProfilePath := cfg.Section("Profile0").Key("Path").String()
	defaultProfilePath = filepath.Join(firefoxDataPath, defaultProfilePath)
	cookiesPath := filepath.Join(defaultProfilePath, "/cookies.sqlite")
	conn, err := sql.Open("sqlite3", cookiesPath+"?mode=ro") // read only
	checkError(err)
	cookies := queryCookiesFirefox(conn)
	return cookies
}
