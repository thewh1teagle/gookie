package main

import (
	"fmt"

	"github.com/thewh1teagle/gookie/gookie"
)

func main() {
	cookies := gookie.Firefox()
	fmt.Println(cookies)
}
