package interact

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/CxZMoE/xz-ease-player/account"
	"github.com/CxZMoE/xz-ease-player/control"
	"github.com/CxZMoE/xz-ease-player/logger"
	xkb "github.com/CxZMoE/xz-keyboard"

	"os"
	"strconv"
	"strings"
	"time"
	//"time"
)

var homedir = ""

type ClientApp struct {
	ShouldClose     bool
	ShouldPrintMenu bool
	BindHandles     []func()
}

type Menu struct {
	Items []string
}

// Play list modes
const (
	MODE_FM     = 0
	MODE_MYFAV  = 1
	MODE_NORMAL = 2
	MODE_INTEL  = 3
	MODE_DAY    = 4
)

func init() {
}

func NewClientApp() *ClientApp {
	c := &ClientApp{ShouldClose: false, ShouldPrintMenu: true}
	return c
}
func fav(player *control.Player, login *account.Login) {
	player.Status = control.StatusOther
	if player.PlayFeature == MODE_MYFAV {
		player.PlayFeature = MODE_NORMAL
	} else {
		player.PlayFeature = MODE_MYFAV
	}
	// Get play sheet struct.
	toIndex := 0
	login.GetAllPlaySheet()
	sheet := login.PlaySheets[toIndex]
	fmt.Printf("\n[SheetName] %s", sheet.Name)
	fmt.Printf("\n[Creator] %s", sheet.CreatorName)

	// Load Detail
	sheet.LoadDetail(login) // Load Song Ids.
	player.PlayFeature = MODE_NORMAL

	player.EmptyPlayList()
	player.Playlist = make([]control.Music, len(sheet.Songs))

	player.NowPlayingSheetId = sheet.Id // For heartbeat mode purpose.

	for i, v := range sheet.Songs {
		fmt.Printf("\n[Song %d] %s", len(sheet.Songs)-i, sheet.Songs[len(sheet.Songs)-i-1].Name)
		player.Playlist[i].Name = v.Name
		player.Playlist[i].PlaySourceType = control.SOURCE_WEB
		player.Playlist[i].Id = v.Id
	}

	if len(player.Playlist) > 0 {
		if player.PlayMode == control.ModeRandom {
			player.SetCurrentIndex(player.GetRandomIndex())
		} else {
			player.SetCurrentIndex(0)
		}
	}
	fmt.Printf("\n[INFO] 切换到我喜欢的音乐")

	//player.Play()
}
func day(player *control.Player, login *account.Login) {
	player.Status = control.StatusOther
	if player.PlayFeature == MODE_DAY {
		player.PlayFeature = MODE_NORMAL
	} else {
		player.PlayFeature = MODE_DAY
	}
	player.EmptyPlayList()
	songs := login.GetRecommend()
	if songs == nil {
		return
	}
	player.Playlist = make([]control.Music, len(songs))
	for i, v := range songs {
		fmt.Printf("\n[Song %d] %s", len(songs)-i, songs[len(songs)-i-1].Name)
		player.Playlist[i].Name = v.Name
		player.Playlist[i].Id = v.Id
		player.Playlist[i].PlaySourceType = control.SOURCE_WEB
	}
	if len(songs) > 0 {
		if player.PlayMode == control.ModeRandom {
			player.SetCurrentIndex(player.GetRandomIndex())
		} else {
			player.SetCurrentIndex(0)
		}
	}
	fmt.Printf("\n[INFO] 切换到每日推荐")
	//player.Play()
}
func fm(player *control.Player, login *account.Login) {
	player.Status = control.StatusOther
	if player.PlayFeature == MODE_FM {
		player.PlayFeature = MODE_NORMAL
	} else {
		player.PlayFeature = MODE_FM
	}

	player.EmptyPlayList()
	songs := login.GetFMSong()
	player.Playlist = make([]control.Music, len(songs))
	for i, v := range songs {
		fmt.Printf("\n[Song %d] %s", len(songs)-i, songs[len(songs)-i-1].Name)
		player.Playlist[i].Id = v.Id
		player.Playlist[i].Name = v.Name
		player.Playlist[i].PlaySourceType = control.SOURCE_WEB
	}
	if len(songs) > 0 {
		if player.PlayMode == control.ModeRandom {
			player.SetCurrentIndex(player.GetRandomIndex())
		} else {
			player.SetCurrentIndex(0)
		}
	}
	fmt.Printf("\n[INFO] 切换到私人FM")
	//player.Play()
}
func qd(player *control.Player, login *account.Login) {
	reqUrl := account.ApiSignIn
	androidQd := login.Client.DoGet(fmt.Sprintf(reqUrl, 0), nil)
	if androidQd == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况")
		return
	}
	if login.GetStatusCode(androidQd) == 200 {
		fmt.Printf("\n[INFO] 安卓登录 +2")
	} else {
		fmt.Printf("\n[INFO] 你已经用安卓登录过一次了")
	}
	pcQd := login.Client.DoGet(fmt.Sprintf(reqUrl, 1), nil)
	if pcQd == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况")
		return
	}
	if login.GetStatusCode(pcQd) == 200 {
		fmt.Printf("\n[INFO] PC/Web 端签到 +1")
	} else {
		fmt.Printf("\n[INFO] 你已经用 PC/Web 端签到过一次了")
	}
}
func stop(player *control.Player) {
	if player.Status == control.StatusPlaying || player.Status == control.StatusPaused || player.Status == control.StatusOther {
		player.Stop()
	}
}

func like(player *control.Player, login *account.Login) {
	nowIndex := player.Playlist[player.GetCurrentIndex()].Id
	if login.Like(nowIndex) {
		fmt.Printf("\n[INFO] 喜欢音乐 [%s] 成功", player.Playlist[player.GetCurrentIndex()].Name)
	} else {
		fmt.Printf("\n[INFO] 喜欢喜欢 [%s] 失败.", player.Playlist[player.GetCurrentIndex()].Name)
	}

}
func mode(player *control.Player) {
	// Play Modes
	const (
		ModeListLoop   = 0
		ModeSingleLoop = 1
		ModeRandom     = 2
		ModeSingleStop = 3
	)
	cm := player.PlayMode
	if cm >= 0 && cm < 3 {
		cm += 1
	}
	if cm >= 3 {
		cm = 0
	}
	player.PlayMode = cm
	cmtext := ""
	switch player.PlayMode {
	case ModeListLoop:
		cmtext = "List Loop"
		break
	case ModeSingleLoop:
		cmtext = "Single Loop"
		break
	case ModeRandom:
		cmtext = "Random"
		break
	case ModeSingleStop:
		cmtext = "Single Stop"
		break
	}
	fmt.Printf("\n[INFO] 切换播放模式到: %s", cmtext)
}

// MainLoop
func (c *ClientApp) MainLoop(login *account.Login, player *control.Player) {
	defer func() {
		//捕获test抛出的panic
		if err := recover(); err != nil {
			fmt.Printf("\n[ERR] 请检查您的网络状况")
			logger.WriteLog(fmt.Sprint(err))
		}
	}()

	fmt.Printf("\n[INFO] 输入 'm' 查看帮助菜单")
	go ShowLyric(player, login, 0)

	// Hot key bindings
	k := xkb.NewKeyboard()
	if k == nil {
		// CTRL+ALT+RightArrow: Next Song
		// CTRL+ALT+LeftArrow: Last Song
		// CTRL+ALT+PgUp: Fast backward
		// CTRL+ALT+PgDn: Fast forward
		// CTRL+ALT+P: Play/Pause
		// CTRL+ALT+]: Increase volume
		// CTRL+ALT+[: Decrease volume
		// CTRL+ALT+F: Go favorite mode
		// CTRL+ALT+G: Go FM mode
		// CTRL+ALT+S: Stop playing
		// CTRL+ALT+D: Go day recommend mode
		// CTRL+ALT+M: Change mode
		// CTRL+ALT+L: Like this song
	} else {
		defer k.StopReadEvent()
		// CTRL+ALT+RightArrow: Next Song
		k.BindKeyEvent("keyboard_next", func() {
			player.Next()
		}, k.Keys[xkb.LCTRL], k.Keys[xkb.LALT], k.Keys[xkb.RIGHT])

		// CTRL+ALT+LeftArrow: Last Song
		k.BindKeyEvent("keyboard_last", func() {
			player.Last()
		}, k.Keys[xkb.LCTRL], k.Keys[xkb.LALT], k.Keys[xkb.LEFT])

		// CTRL+ALT+P: Play/Pause
		k.BindKeyEvent("keyboard_pause", func() {
			if player.Status == control.StatusPlaying {
				player.Pause()
			} else {
				player.Play()
			}
		}, k.Keys[xkb.LCTRL], k.Keys[xkb.LALT], k.Keys[xkb.P])

		// CTRL+ALT+]: Increase volume
		k.BindKeyEvent("keyboard_vol_increase", func() {
			vol := player.GetVolume() + 10
			player.SetVolume(vol)
			fmt.Printf("\n[VOL] %d%%", vol)
		}, k.Keys[xkb.LCTRL], k.Keys[xkb.LALT], k.Keys[xkb.Zkhy])

		// CTRL+ALT+[: Decrease volume
		k.BindKeyEvent("keyboard_vol_decrease", func() {
			vol := player.GetVolume() - 10
			player.SetVolume(vol)
			fmt.Printf("\n[VOL] %d%%", vol)
		}, k.Keys[xkb.LCTRL], k.Keys[xkb.LALT], k.Keys[xkb.Zkhz])

		// CTRL+ALT+F: Go favorite mode
		k.BindKeyEvent("keyboard_fav", func() {
			fav(player, login)
		}, k.Keys[xkb.LCTRL], k.Keys[xkb.LALT], k.Keys[xkb.F])

		// CTRL+ALT+G: Go FM mode
		k.BindKeyEvent("keyboard_fm", func() {
			fm(player, login)
		}, k.Keys[xkb.LCTRL], k.Keys[xkb.LALT], k.Keys[xkb.G])

		// CTRL+ALT+S: Stop playing
		k.BindKeyEvent("keyboard_stop", func() {
			stop(player)
		}, k.Keys[xkb.LCTRL], k.Keys[xkb.LALT], k.Keys[xkb.S])

		// CTRL+ALT+D: Go day recommend mode
		k.BindKeyEvent("keyboard_day", func() {
			day(player, login)
		}, k.Keys[xkb.LCTRL], k.Keys[xkb.LALT], k.Keys[xkb.D])

		// CTRL+ALT+M: Change mode
		k.BindKeyEvent("keyboard_mode", func() {
			mode(player)
		}, k.Keys[xkb.LCTRL], k.Keys[xkb.LALT], k.Keys[xkb.M])

		// CTRL+ALT+L: Like this song
		k.BindKeyEvent("keyboard_like", func() {
			like(player, login)
		}, k.Keys[xkb.LCTRL], k.Keys[xkb.LALT], k.Keys[xkb.L])

		// CTRL+ALT+PgUp: Fast backward
		k.BindKeyEvent("keyboard_fastforward", func() {
			var secInt = player.Playlist[player.GetCurrentIndex()].GetPosition() - 15
			length := player.Playlist[player.GetCurrentIndex()].GetLength()
			if secInt > length || secInt < 0 {

			} else {
				player.Playlist[player.GetCurrentIndex()].SetPosition(secInt)
				go func() {
					if !player.IsShowProgress {
						player.IsShowProgress = true
						time.Sleep(time.Millisecond * 200)
						player.IsShowProgress = false
					}
				}()
			}
		}, k.Keys[xkb.LCTRL], k.Keys[xkb.LALT], k.Keys[xkb.PgUp])

		// CTRL+ALT+PgDn: Fast forward
		k.BindKeyEvent("keyboard_fastbackward", func() {
			var secInt = player.Playlist[player.GetCurrentIndex()].GetPosition() + 15
			length := player.Playlist[player.GetCurrentIndex()].GetLength()
			if secInt > length || secInt < 0 {

			} else {
				player.Playlist[player.GetCurrentIndex()].SetPosition(secInt)
				go func() {
					if !player.IsShowProgress {
						player.IsShowProgress = true
						time.Sleep(time.Millisecond * 200)
						player.IsShowProgress = false
					}
				}()
			}
		}, k.Keys[xkb.LCTRL], k.Keys[xkb.LALT], k.Keys[xkb.PgDn])
	}

	//defer k.StopReadEvent()

	//go ShowLyric(player,login)
	for !c.ShouldClose {
		c.PrintPrefix(">> ")
		input := getInput()
		c.inputHandler(login, player, input)
	}
}

func (c *ClientApp) inputHandler(login *account.Login, player *control.Player, input string) {
	args := strings.Split(input, " ")
	cmd := args[0]
	switch cmd {
	case "author":
		fmt.Printf("\n[Author] CxZMoE\n[EMAIL] forevermisaka@163.com\n[GITHUB] https://github.com/CxZMoE\n[MSG] If you find a bug,please contact me,thanks.\n[MSG] This software is free,if you have paid for this,you were cheated.")
		break
	case "login":
		email, passwd := args[1], args[2]
		//fmt.Printf("\n[INFO] Login with:")
		fmt.Printf("\n[INFO] 使用邮箱: %s 进行登录", email)
		//fmt.Printf("\n[INFO] Password: %s", passwd)
		login.LoginEmail(email, passwd)
		if login.LoginData != nil {
			fmt.Printf("\n[INFO] 登录成功")
		}
		break
	case "logout":
		if login.Client.LoginStatus == true {
			data := login.Logout()
			fmt.Printf("\n%s", string(data))
			homedir, err := os.UserHomeDir()
			if err != nil {
				logger.WriteLog("Failed to get home path.")
				return
			}
			os.Remove(homedir + "/xzp/cookie")
		} else {
			fmt.Printf("\nY你还没有登录")
		}
		break
	case "fm":
		fm(player, login)
		break
	case "fav":
		fav(player, login)
		break
	case "like":
		like(player, login)
		break
	case "day":
		day(player, login)
		break
	case "sheet":
		player.Status = control.StatusOther
		// Get play sheet struct.
		login.GetAllPlaySheet()
		if len(args) == 1 {
			fmt.Printf("\n===歌单===")
			for i, _ := range login.PlaySheets {
				fmt.Printf("\n[%d] %s [%s]", len(login.PlaySheets)-i, login.PlaySheets[len(login.PlaySheets)-i-1].Name, login.PlaySheets[len(login.PlaySheets)-i-1].CreatorName)
			}
			fmt.Printf("\n============")
		} else if len(args) == 2 {
			toIndex, err := strconv.Atoi(args[1])
			if err != nil {
				logger.WriteLog(err.Error())
				break
			}
			login.GetAllPlaySheet()
			sheet := login.PlaySheets[toIndex]
			fmt.Printf("\n[歌单] %s", sheet.Name)
			fmt.Printf("\n[创建者] %s", sheet.CreatorName)

			// Load Detail
			sheet.LoadDetail(login) // Load Song Ids.
			player.PlayFeature = MODE_NORMAL

			player.EmptyPlayList()
			player.Playlist = make([]control.Music, len(sheet.Songs))

			player.NowPlayingSheetId = sheet.Id // For heartbeat mode purpose.

			for i, v := range sheet.Songs {
				fmt.Printf("\n[歌曲 %d] %s", len(sheet.Songs)-i, sheet.Songs[len(sheet.Songs)-i-1].Name)
				player.Playlist[i].Name = v.Name
				player.Playlist[i].PlaySourceType = control.SOURCE_WEB
				player.Playlist[i].Id = v.Id
			}

			if len(player.Playlist) > 0 {
				if player.PlayMode == control.ModeRandom {
					player.SetCurrentIndex(player.GetRandomIndex())
				} else {
					player.SetCurrentIndex(0)
				}
			}
			//player.Play()
			fmt.Printf("\n[INFO] 切换播放列表到歌单 %s", sheet.Name)
		}
		break
	case "list":
		if len(player.Playlist) > 0 {
			for i, _ := range player.Playlist {
				fmt.Printf("\n[%d] %s", len(player.Playlist)-i, player.Playlist[len(player.Playlist)-i-1].Name)
			}
		}
		break
	case "ls":
		if len(player.Playlist) > 0 {
			for i, _ := range player.Playlist {
				fmt.Printf("\n[%d] %s", len(player.Playlist)-i, player.Playlist[len(player.Playlist)-i-1].Name)
			}
		}
		break
	case "n":
		player.Next()
		break
	case "next":
		player.Next()
		break
	case "l":
		player.Last()
		break
	case "last":
		player.Next()
		break
	case "play":
		if len(player.Playlist) > 0 {
			player.Play()
		} else {
			fmt.Printf("\n[INFO] 播放列表为空")
		}
		break
	case "p":
		if len(player.Playlist) > 0 {
			player.Play()
		} else {
			fmt.Printf("\n[INFO] 播放列表为空")
		}
		break
	case "pause":
		if player.Status == control.StatusPlaying {
			player.Pause()
		}
		break
	case "stop":
		stop(player)
		break
	case "goto":
		index, _ := strconv.Atoi(args[1])
		index -= 1
		if index < 0 || index >= len(player.Playlist) {
			fmt.Printf("\n[INFO] goto: 序号超出范围")
			break
		}
		player.Stop()
		//player.FreeStream(player.Playlist[player.GetCurrentIndex()].Handle)
		player.SetCurrentIndex(index)
		player.Status = control.StatusPlaying
		player.Play()
		break
	case "go":
		index, _ := strconv.Atoi(args[1])
		index -= 1
		if index < 0 || index >= len(player.Playlist) {
			fmt.Printf("\n[INFO] goto: 序号超出范围")
			break
		}
		player.Stop()
		//player.FreeStream(player.Playlist[player.GetCurrentIndex()].Handle)
		player.SetCurrentIndex(index)
		player.Status = control.StatusPlaying
		player.Play()
		break
	case "pg":
		player.IsShowProgress = !player.IsShowProgress
		break
	case "i":
		player.Status = control.StatusOther
		if player.PlayFeature != MODE_NORMAL {
			logger.WriteLog("Please choose a play sheet first,Use: sheet [id]")
		} else {
			player.PlayFeature = MODE_INTEL
			//log.Println(fmt.Sprintf(account.ApiHeartbeatMode,player.Playlist[player.GetCurrentIndex()].Id,player.NowPlayingSheetId))
			reqUrl := fmt.Sprintf(account.ApiHeartbeatMode, player.Playlist[player.GetCurrentIndex()].Id, player.NowPlayingSheetId)
			listDatas := login.Client.DoGet(reqUrl, nil)
			if listDatas == nil {
				fmt.Printf("\n[ERR] 请检查您的网络状况")
				break
			}
			var listDatasJson interface{}
			var ss []account.Song
			json.Unmarshal(listDatas, &listDatasJson)
			data := listDatasJson.(map[string]interface{})["data"]
			if data != nil {
				for _, v := range data.([]interface{}) {
					var s account.Song
					songInfo := v.(map[string]interface{})["songInfo"]
					if songInfo != nil {
						id := songInfo.(map[string]interface{})["id"]
						name := songInfo.(map[string]interface{})["name"]
						//log.Println(name)
						if id == nil || name == nil {
							continue
						}
						s.Id = int(id.(float64))
						s.Name = name.(string)

					} else {
						continue
					}
					ss = append(ss, s)
				}
			} else {
				break
			}

			//if player.Status == control.StatusPlaying || player.Status == control.StatusPaused || player.Status == control.StatusOther {
			//	player.Stop()
			//player.FreeStream(player.Playlist[player.GetCurrentIndex()].Handle)
			//}

			player.EmptyPlayList()
			//log.Println("Length:",len(ss))
			player.Playlist = make([]control.Music, len(ss))
			for i, v := range ss {
				player.Playlist[i].Name = v.Name
				player.Playlist[i].Id = v.Id
				player.Playlist[i].PlaySourceType = control.SOURCE_WEB
			}

			if len(player.Playlist) > 0 {
				if player.PlayMode == control.ModeRandom {
					player.SetCurrentIndex(player.GetRandomIndex())
				} else {
					player.SetCurrentIndex(0)
				}
			}
			fmt.Printf("\n[INFO] Entered heartbeat mode.")
			//player.Play()
		}
		break
	case "mode":
		mode(player)
		break
	case "qd":
		qd(player, login)
		break
	case "m":
		menu := Menu{Items: []string{}}
		fmt.Printf("\n===指令列表===")
		menu.AddItem("[author] 显示作者信息")
		menu.AddItem("[login] <邮箱> <密码>: 邮箱登陆")
		menu.AddItem("[logout]: 登出")
		menu.AddItem("[qd]: 每日签到")
		menu.AddItem("[fm]: 前往私人FM模式")
		menu.AddItem("[fav]: 前往我喜欢的音乐")
		menu.AddItem("[day]: 前往每日推荐")
		menu.AddItem("[sheet]: 显示当前歌单列表")
		menu.AddItem("[sheet] <序号>: 前往对应序号歌单")
		menu.AddItem("[list/ls]: 显示播放列表")
		menu.AddItem("[goto/go] <序号>: 转跳到指定序号歌曲")
		menu.AddItem("[time/t] <sec>: 跳到歌曲的第sec秒")
		menu.AddItem("[last/l]: 上一首")
		menu.AddItem("[next/n]: 下一首")
		menu.AddItem("[play/p]: 播放歌曲")
		menu.AddItem("[pause]: 暂停歌曲")
		menu.AddItem("[stop]: 停止歌曲")
		menu.AddItem("[pg]: 显示进度条 #显示的时候输入字符会被刷掉.")
		menu.AddItem("[key] 显示快捷键列表")
		c.PrintMenu(menu)
		fmt.Printf("\n====================")
		break
	case "vol":
		value, err := strconv.Atoi(args[1])
		if value < 0 || value > 100 {
			fmt.Printf("\n[INFO] 错误的值(过大/过小)")
			break
		}
		if err != nil {
			fmt.Printf("\n[INFO] 错误的值")
			break
		}
		if player.SetVolume(uint(value)) == 1 {
			fmt.Printf("\n[INFO] 设置音量为 %d", value)
			break
		} else {
			fmt.Printf("\n[INFO] 设置音量失败")
			break
		}

		break
	case "x": // Lyric ON/OFF
		if player.LyricSwitch == true {
			player.LyricSwitch = false
			fmt.Printf("\n[INFO] 歌词关闭")
			break
		}
		player.LyricSwitch = true
		fmt.Printf("\n[INFO] 歌词开启")
		break
	case "key":
		fmt.Printf("\n[INFO] 快捷键列表\n")
		fmt.Printf(`
		CTRL+ALT+右箭头: 下一首
		CTRL+ALT+左箭头: 上一首
		CTRL+ALT+PgDn: 快进 15s
		CTRL+ALT+PgUp: 快退 15s
		CTRL+ALT+P: 播放/暂停
		CTRL+ALT+S: 停止播放
		CTRL+ALT+]: 增加音量 10%%
		CTRL+ALT+[: 减少音量 10%%
		CTRL+ALT+F: 前往我喜欢的音乐
		CTRL+ALT+G: 前往私人FM
		CTRL+ALT+D: 前往推荐
		CTRL+ALT+M: 改变播放模式
		CTRL+ALT+L: 添加到喜欢`)
	case "exit":
		control.Release()
		c.ShouldClose = true
		break
	case "t":
		sec := args[1]
		secInt, err := strconv.Atoi(sec)
		if err != nil {
			fmt.Printf("\n[ERR] 错误的值")
			break
		}
		length := player.Playlist[player.GetCurrentIndex()].GetLength()
		if secInt > length || secInt < 0 {
			fmt.Printf("\n[ERR] 值 %d 超出 0-%d 秒的范围", secInt, length)
			break
		}
		player.Playlist[player.GetCurrentIndex()].SetPosition(secInt)
		go func() {
			if !player.IsShowProgress {
				player.IsShowProgress = true
				time.Sleep(time.Millisecond * 200)
				player.IsShowProgress = false
			}
		}()
		break
	case "time":
		sec := args[1]
		secInt, err := strconv.Atoi(sec)
		if err != nil {
			fmt.Printf("\n[ERR] 错误的值")
			break
		}
		length := player.Playlist[player.GetCurrentIndex()].GetLength()
		if secInt > length || secInt < 0 {
			fmt.Printf("\n[ERR] 值 %d 超出 0-%d 秒的范围", secInt, length)
			break
		}
		player.Playlist[player.GetCurrentIndex()].SetPosition(secInt)
		go func() {
			if !player.IsShowProgress {
				player.IsShowProgress = true
				time.Sleep(time.Millisecond * 200)
				player.IsShowProgress = false
			}
		}()
		break
	}

}

func getInput() string {
	reader := bufio.NewReader(os.Stdin)
	input, _, _ := reader.ReadLine()
	return string(input)
}
func (c *ClientApp) PrintPrefix(prefix string) {
	fmt.Printf("\n%s", prefix)
}

func (c *ClientApp) PrintMenu(menu Menu) {
	if !c.ShouldPrintMenu {
		return
	}
	for i, v := range menu.Items {
		fmt.Printf("\n%d) %s", i, v)
	}
}

func (m *Menu) AddItem(value string) *Menu {
	if m.Items != nil {
		m.Items = append(m.Items, value)
		return m
	} else {
		fmt.Printf("\n[INFO] 菜单项目不应为空")
		return nil
	}
}

func ShowLyric(player *control.Player, login *account.Login, offset int) {
	if err := recover(); err != nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况")
		logger.WriteLog(fmt.Sprint(err))
	}
	for {
		time.Sleep(time.Second * 1)
		if player.LyricSwitch == false {
			continue
		}
		if player.Status == control.StatusPlaying || player.Status == control.StatusPaused || player.Status == control.StatusOther {
		RELYRIC:
			lyric := GetLyricCurrent(player, login)
			if lyric == "" {
				fmt.Printf("\n[ERR] 请检查您的网络状况")
				return
			}
			lastContent := ""
			lastName := player.Playlist[player.GetCurrentIndex()].Name
			parsed := ParseLyric(lyric)
			for {
				if lastName != player.Playlist[player.GetCurrentIndex()].Name {
					goto RELYRIC
				}
				if player.LyricSwitch == false {
					break
				}

				line := parsed[player.Playlist[player.GetCurrentIndex()].GetPosition()+offset]
				if line.Content == "" {

				} else {
					content := line.Content
					if lastContent != content {
						fmt.Printf("\n[L] %s", content)
						lastContent = content
					}
				}

				time.Sleep(time.Millisecond * 500)
			}
		}
	}
}

func GetLyricCurrent(player *control.Player, login *account.Login) string {
	id := player.Playlist[player.GetCurrentIndex()].Id
	reqUrl := fmt.Sprintf(account.ApiLyric, id)
	lyricData := login.Client.DoGet(reqUrl, nil)
	if lyricData == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况")
		return ""
	}
	var dataJson interface{}
	json.Unmarshal(lyricData, &dataJson)
	if dataJson != nil {
		lrc := dataJson.(map[string]interface{})["lrc"]
		if lrc != nil {
			lyric := lrc.(map[string]interface{})["lyric"]
			if lyric != nil {
				return lyric.(string)
			} else {
				return ""
			}
		} else {
			return ""
		}
	} else {
		return ""
	}
}

type LyricLine struct {
	Min     int
	Sec     int
	Content string
	Length  int
}

func ParseLyric(src string) map[int]LyricLine {
	var lls []LyricLine
	var llss map[int]LyricLine
	//fmt.Println(src)
	lines := strings.Split(src, "\n")
	// [00:12.570]难以忘记初次见你
	for _, v := range lines {
		var l LyricLine
		//fmt.Println(v)
		unit := strings.TrimLeft(v, "[")
		// 00:12.570]难以忘记初次见你
		value := " "
		if len(strings.Split(unit, "]")) > 1 {
			value = strings.Split(unit, "]")[1]
		}
		// 难以忘记初次见你

		times := strings.Split(unit, ".")[0]
		// 00:12
		if len(times) == 0 {
			continue
		}
		timeMin := strings.TrimPrefix(strings.Split(times, ":")[0], "0")
		// 0
		timeMinInt, err := strconv.Atoi(timeMin)
		//fmt.Println(times)
		if err != nil {
			logger.WriteLog(err.Error())
		}
		timeSec := strings.TrimPrefix(strings.Split(times, ":")[1], "0")
		// 12
		timeSecInt, err := strconv.Atoi(timeSec)
		if err != nil {
			logger.WriteLog(err.Error())
		}
		l.Min, l.Sec, l.Content = timeMinInt, timeSecInt, value
		l.Length = l.Min*60 + l.Sec
		lls = append(lls, l)
	}

	llss = make(map[int]LyricLine, len(lls))
	for _, v := range lls {
		llss[v.Length] = v
	}
	return llss
}
