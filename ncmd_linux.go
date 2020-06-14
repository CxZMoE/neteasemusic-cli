// +build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/CxZMoE/NetEase-CMD/account"
	"github.com/CxZMoE/NetEase-CMD/control"
	"github.com/CxZMoE/NetEase-CMD/interact"
	"github.com/CxZMoE/NetEase-CMD/logger"
	"github.com/CxZMoE/NetEase-CMD/network"
)

var me account.Login
var apiPid int

func init() {
	chdir()
	// Load NetEase Api
	apiPid = loadAPI()

}

// StartAPI 启动网易云API
func StartAPI() *exec.Cmd {
	//homedir, err := os.UserHomeDir()
	//if err != nil {
	//	logger.WriteLog("无法获取用户目录地质")
	//	return nil
	//}

	apiExecPath := "/usr/share/NetEase-CMD/NeteaseApi/app.js"
	_, err := os.Stat(apiExecPath)
	if os.IsNotExist(err) {
		// 没有API文件不能启动程序
		logger.WriteLog("Couldn't start API server,app.js not found.")
		panic(err)
	}
	cmd := exec.Command("node", apiExecPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	//output,_ := cmd.Output()
	//log.Println(output)
	return cmd
}

// AppRun run application
func AppRun() {
	defer func() {
		//捕获test抛出的panic
		if err := recover(); err != nil {
			fmt.Printf("\n[ERR] AR: 请检查您的网络状况")
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
	_, err = os.Stat(homedir + "/.ncmd/cookie")
	if os.IsNotExist(err) {

		fmt.Printf("\n[INFO] 你还没有登录呢,请先登录")
	} else {
		client.LoadJar(homedir + "/.ncmd/cookie")
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

// 加载API
func loadAPI() int {
	loadedAPI := make(chan bool, 1)
	result := StartAPI()
	if result.Process != nil {
		logger.WriteLog("API Started.")
		loadedAPI <- true
	} else {
		panic(result)
	}

	<-loadedAPI
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
