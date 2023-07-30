package main

import (
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func checkError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func decryptEncryptedValue(value []byte, key []byte) string {
	value = value[3:]
	nonce := value[:12]
	// tag := encryptedValue[len(encryptedValue)-16:]
	value = value[12:]
	block, err := aes.NewCipher(key)
	checkError(err)
	aesgcm, err := cipher.NewGCM(block)
	checkError(err)

	decrypted, err := aesgcm.Open(nil, nonce, value, nil)
	if err != nil {
		panic(err.Error())
	}
	strValue := string(decrypted)
	return strValue
}

type RawCookie struct {
	HostKey            string `json:"hostKey"`
	Path               string `json:"path"`
	IsSecure           int    `json:"isSecure"`
	ExpiresNtTimeEpoch int    `json:"expiresNtTimeEpoch"`
	Name               string `json:"name"`
	Value              string `json:"value"`
	IsHttpOnly         int    `json:"isHttpOnly"`
	SameSite           int    `json:"sameSite"`
}

func queryCookies(conn *sql.DB, v10Key []byte, optionalParams ...string) []RawCookie {
	var domain string = "" // default value
	if len(optionalParams) > 0 {
		domain = optionalParams[0]
	}
	query := `
	SELECT host_key, path, is_secure, expires_utc, name, value, encrypted_value, is_httponly, samesite
	FROM cookies
	WHERE host_key like ?;
	`

	rows, err := conn.Query(query, "%"+domain+"%")
	if err != nil {
		// If the first query fails, replace the string and try the second query
		query = strings.Replace(query, "is_secure", "secure", 1)
		rows, err = conn.Query(query)
		checkError(err)
	}
	defer rows.Close()
	defer conn.Close()

	var cookies []RawCookie

	for rows.Next() {
		var hostKey, path, name, value string
		var isSecure, isHttpOnly, sameSite, expiresNtTimeEpoch int
		var encryptedValue []byte

		err := rows.Scan(&hostKey, &path, &isSecure, &expiresNtTimeEpoch, &name, &value, &encryptedValue, &isHttpOnly, &sameSite)
		if err != nil {
			fmt.Println("Error scanning row:", err)
			continue
		}
		value = decryptEncryptedValue(encryptedValue, v10Key)
		cookie := RawCookie{
			HostKey:            hostKey,
			Path:               path,
			IsSecure:           isSecure,
			ExpiresNtTimeEpoch: expiresNtTimeEpoch,
			Name:               name,
			Value:              value,
			IsHttpOnly:         isHttpOnly,
			SameSite:           sameSite,
		}
		cookies = append(cookies, cookie)
	}
	return cookies
}

func getChromeCookies(keyPath string, cookiesPath string) []RawCookie {
	keyContent, _ := ioutil.ReadFile(keyPath)
	var jsonKey map[string]interface{}
	err := json.Unmarshal(keyContent, &jsonKey)
	checkError(err)
	key64 := jsonKey["os_crypt"].(map[string]interface{})["encrypted_key"].(string)
	keydpapi, err := base64.StdEncoding.DecodeString(key64)
	checkError(err)
	keydpapi = keydpapi[5:]
	v10key, err := Decrypt(keydpapi)
	checkError(err)

	conn, err := sql.Open("sqlite3", cookiesPath+"?mode=ro") // read only
	checkError(err)
	return queryCookies(conn, v10key)
}

func findChromePaths() (string, string) {
	appDataPath := os.Getenv("appdata")
	userDataPath := fmt.Sprintf("%v%v", appDataPath, filepath.FromSlash("/../local/Google/Chrome/User Data"))
	keyPath := filepath.Join(userDataPath, "Local State")
	dbPath := filepath.Join(userDataPath, "Default/Network/Cookies")
	return keyPath, dbPath
}

func main() {
	keyPath, dbPath := findChromePaths()
	cookies := getChromeCookies(keyPath, dbPath)
	for _, cookie := range cookies {
		fmt.Println(cookie)
	}
}
