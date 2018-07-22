// +build windows

package manageusers

import (
	"fmt"
)

func doReload() {
	fmt.Println("Can't reload Caddy on Windows.")
}
