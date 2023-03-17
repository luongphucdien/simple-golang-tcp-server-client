// THIS IS FOR TESTING PURPOSE ONLY //

package main

import (
	"fmt"
	"strings"
)

// import tcpserver "github.com/edibl/go/tcp-server"

func main() {
	var arr []string
	for i := 0; i< 3; i++ {
		arr = append(arr, "_*")
	}
	// for i := range(arr) {
	// 	copy(arr[i:], []string{"o"})
	// }
	str := strings.Join(arr[:], "")
	fmt.Print(str)
}
