# gookie

`GO` version for [browser_cookie3](https://github.com/borisbabic/browser_cookie3)

* ***What does it do?*** Loads cookies used by your web browser into a `cookiejar` object / `json`.
* ***Why is it useful?*** This means you can use `go` to download and get the same content you see in the web browser without needing to login.
* ***Which browsers are supported?*** `Chrome` ,`Firefox`, ~~`LibreWolf`, `Opera`, `Opera GX`, `Edge`, `chromium`, `Brave`, `Vivaldi`, and `Safari`~~.
* ***How are the cookies stored?*** All currently-supported browsers store cookies in a `sqlite` database in your home directory.

### Install
```shell
go get github.com/thewh1teagle/gookie/gookie
```
### Usage
```go
package main

import (
	"net/http"
	"github.com/thewh1teagle/gookie/gookie"
)

func main() {
    cookies := gookie.Chrome()
    jar := gookie.ToCookieJar(cookies)

    client := &http.Client{ // Now your session has chrome cookies
        Jar: jar,
    }
    
}
```

### TODO

- [x] Support Firefox
- [ ] Support Linux / OSX (currenly only firefox supported)
- [ ] Cross Platform code organize (build constraints)
- [ ] Makefile
- [x] Examples
- [x] Create go package at pkg.go.dev
- [x] Load into `net/http` with `cookiejar`
- [ ] Tests