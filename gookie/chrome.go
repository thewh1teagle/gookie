package gookie

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
)

func getChromeCookies(keyPath string, cookiesPath string) []Cookie {
	keyContent, _ := ioutil.ReadFile(keyPath)
	var jsonKey map[string]interface{}
	err := json.Unmarshal(keyContent, &jsonKey)
	checkError(err)
	key64 := jsonKey["os_crypt"].(map[string]interface{})["encrypted_key"].(string)
	keydpapi, err := base64.StdEncoding.DecodeString(key64)
	checkError(err)
	keydpapi = keydpapi[5:]
	v10key, err := decrypt(keydpapi)
	checkError(err)

	conn, err := sql.Open("sqlite3", cookiesPath+"?mode=ro") // read only
	checkError(err)
	return queryCookiesChrome(conn, v10key)
}

func findChromePaths() (string, string) {
	appDataPath := os.Getenv("APPDATA")
	userDataPath := fmt.Sprintf("%v%v", appDataPath, filepath.FromSlash("/../local/Google/Chrome/User Data"))
	keyPath := filepath.Join(userDataPath, "Local State")
	dbPath := filepath.Join(userDataPath, "Default/Network/Cookies")
	return keyPath, dbPath
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

func queryCookiesChrome(conn *sql.DB, v10Key []byte, optionalParams ...string) []Cookie {
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

	var cookies []Cookie

	for rows.Next() {
		var hostKey, path, name, value string
		var sameSite, expiresNtTimeEpoch int
		var isSecure, isHttpOnly bool
		var encryptedValue []byte

		err := rows.Scan(&hostKey, &path, &isSecure, &expiresNtTimeEpoch, &name, &value, &encryptedValue, &isHttpOnly, &sameSite)
		if err != nil {
			fmt.Println("Error scanning row:", err)
			continue
		}

		expires := expiresNtTimeEpochToTime(int64(expiresNtTimeEpoch))
		value = decryptEncryptedValue(encryptedValue, v10Key)
		cookie := Cookie{
			Host:       hostKey,
			Path:       path,
			IsSecure:   isSecure,
			Expires:    expires,
			Name:       name,
			Value:      value,
			IsHttpOnly: isHttpOnly,
			SameSite:   sameSite,
		}
		cookies = append(cookies, cookie)
	}
	return cookies
}

func Chrome(params ...string) []Cookie {
	var keyPath, dbPath string

	// If parameters are provided, use them directly
	if len(params) >= 2 {
		keyPath = params[0]
		dbPath = params[1]
	} else {
		// If parameters are not provided, get them from findChromePaths function
		keyPath, dbPath = findChromePaths()
	}

	cookies := getChromeCookies(keyPath, dbPath)
	return cookies
}
