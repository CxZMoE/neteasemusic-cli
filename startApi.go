package main

import (
	"github.com/CxZMoE/xz-ease-player/logger"
	"os"
	"os/exec"
)

func main() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		logger.WriteLog("Failed to get home path")
		return
	}
	cmd := exec.Command("node", homedir+"/xzp/NeteaseApi/app.js")
	cmd.Start()
}
