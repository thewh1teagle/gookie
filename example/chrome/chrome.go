package main

import (
	"fmt"

	"github.com/thewh1teagle/gookie/gookie"
)

func main() {
	cookies := gookie.Chrome()
	fmt.Println(cookies)
}
