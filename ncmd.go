package main

import (
	"fmt"
	"github.com/CxZMoE/xz-ease-player/account"
	"github.com/CxZMoE/xz-ease-player/control"
	"github.com/CxZMoE/xz-ease-player/interact"
	"github.com/CxZMoE/xz-ease-player/logger"
	"github.com/CxZMoE/xz-ease-player/network"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

var me account.Login
var apiPid int

func init() {
	chdir()
	// Load NetEase Api
	apiPid = LoadApi()

}

func main() {
	defer func() {
		//捕获test抛出的panic
		if err := recover(); err != nil {
			fmt.Printf("\n[ERR] 请检查您的网络状况")
			logger.WriteLog(fmt.Sprint(err))
		}
	}()
	homedir, err := os.UserHomeDir()
	if err != nil {
		logger.WriteLog("Failed to get home path")
		return
	}
	defer syscall.Kill(-apiPid, syscall.SIGKILL)
	app := interact.NewClientApp()
	client := network.NewClient()
	_, err = os.Stat(homedir + "/xzp/cookie")
	if os.IsNotExist(err) {
		fmt.Printf("\n[INFO] 你还没有登录呢,请先登录")
	} else {
		client.LoadJar(homedir + "/xzp/cookie")
		fmt.Printf("\n[INFO] 已登录")
		fmt.Printf("\n[INFO] 如果无法获取到准确信息,请重新登陆.")
	}
	login := account.NewLogin(client)

	player := control.NewPlayer(login)
	player.SetPlayMode(control.ModeListLoop)
	// Auto Play Control
	go func() {
		for {
			if len(player.Playlist) <= 0 {
				goto SKIPMODE
			}
			if (player.Playlist[player.GetCurrentIndex()].GetLength() <= player.Playlist[player.GetCurrentIndex()].GetPosition()) && player.Playlist[player.GetCurrentIndex()].GetPosition() > 0 {
				switch player.PlayMode {
				case control.ModeListLoop:
					player.Next()
					break
				case control.ModeSingleLoop:
					player.Play()
					break
				case control.ModeRandom:
					player.Next()
					break
				case control.ModeSingleStop:
					player.Stop()
					break
				default:
					break
				}
			}
		SKIPMODE:
			time.Sleep(time.Second * 1)
		}

	}()

	app.MainLoop(login, player)
}

func LoadApi() int {
	loadedApi := make(chan bool, 1)
	result := control.StartAPI()
	if result.Process != nil {
		logger.WriteLog("API Started.")
		loadedApi <- true
	} else {
		panic(result)
	}

	<-loadedApi
	return result.Process.Pid
}

func chdir() (err error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return
	}

	err = os.Chdir(dir)
	return
}
