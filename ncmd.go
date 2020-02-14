package main

import (
	"fmt"
	"runtime"
)

func main() {
	os := runtime.GOOS
	switch os {
	case "windows":
		AppRun()
		break
	case "linux":
		AppRun()
		break
	default:
		fmt.Println("[ERR] Unknown system type.")
	}
}
