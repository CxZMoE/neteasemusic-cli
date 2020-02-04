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
)

var homedir = ""

// ClientApp 网易云客户端
type ClientApp struct {
	ShouldClose     bool
	ShouldPrintMenu bool
	BindHandles     []func()
}

// Menu 帮助菜单
type Menu struct {
	Items []string
}

// Play list modes
const (
	ModeFM     = 0
	ModeMyFav  = 1
	ModeNormal = 2
	ModeIntel  = 3
	ModeDay    = 4
)

func init() {
}

// NewClientApp 创建新网易云客户端
func NewClientApp() *ClientApp {
	c := &ClientApp{ShouldClose: false, ShouldPrintMenu: true}
	return c
}
func fav(player *control.Player, login *account.Login) {
	player.Status = control.StatusOther
	if player.PlayFeature == ModeMyFav {
		player.PlayFeature = ModeNormal
	} else {
		player.PlayFeature = ModeMyFav
	}
	// Get play sheet struct.
	toIndex := 0
	login.GetAllPlaySheet()
	sheet := login.PlaySheets[toIndex]
	fmt.Printf("\n[SheetName] %s", sheet.Name)
	fmt.Printf("\n[Creator] %s", sheet.CreatorName)

	// Load Detail
	sheet.LoadDetail(login) // Load Song IDs.
	player.PlayFeature = ModeNormal

	player.EmptyPlayList()
	player.Playlist = make([]control.Music, len(sheet.Songs))

	player.NowPlayingSheetID = sheet.ID // For heartbeat mode purpose.

	for i, v := range sheet.Songs {
		fmt.Printf("\n[Song %d] %s", len(sheet.Songs)-i, sheet.Songs[len(sheet.Songs)-i-1].Name)
		player.Playlist[i].Name = v.Name
		player.Playlist[i].PlaySourceType = control.SourceWeb
		player.Playlist[i].ID = v.ID
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
	if player.PlayFeature == ModeDay {
		player.PlayFeature = ModeNormal
	} else {
		player.PlayFeature = ModeDay
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
		player.Playlist[i].ID = v.ID
		player.Playlist[i].PlaySourceType = control.SourceWeb
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
	if player.PlayFeature == ModeFM {
		player.PlayFeature = ModeNormal
	} else {
		player.PlayFeature = ModeFM
	}

	player.EmptyPlayList()
	songs := login.GetFMSong()
	player.Playlist = make([]control.Music, len(songs))
	for i, v := range songs {
		fmt.Printf("\n[Song %d] %s", len(songs)-i, songs[len(songs)-i-1].Name)
		player.Playlist[i].ID = v.ID
		player.Playlist[i].Name = v.Name
		player.Playlist[i].PlaySourceType = control.SourceWeb
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
	reqURL := account.APISignIn
	androidQD := login.Client.DoGet(fmt.Sprintf(reqURL, 0), nil)
	if androidQD == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况")
		return
	}
	if login.GetStatusCode(androidQD) == 200 {
		fmt.Printf("\n[INFO] 安卓登录 +2")
	} else {
		fmt.Printf("\n[INFO] 你已经用安卓登录过一次了")
	}
	pcQd := login.Client.DoGet(fmt.Sprintf(reqURL, 1), nil)
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
	nowIndex := player.Playlist[player.GetCurrentIndex()].ID
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
		cm++
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

// MainLoop 客户端主循环
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
		fmt.Printf("\n[Author] CxZMoE\n[EMAIL] forevermisaka@163.com\n[GITHUB] https://github.com/CxZMoE\n[MSG] If you find a bug,please contact me,thanks.\n[MSG] This software is free,if you have paID for this,you were cheated.")
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
			for i := range login.PlaySheets {
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
			sheet.LoadDetail(login) // Load Song IDs.
			player.PlayFeature = ModeNormal

			player.EmptyPlayList()
			player.Playlist = make([]control.Music, len(sheet.Songs))

			player.NowPlayingSheetID = sheet.ID // For heartbeat mode purpose.

			for i, v := range sheet.Songs {
				fmt.Printf("\n[歌曲 %d] %s", len(sheet.Songs)-i, sheet.Songs[len(sheet.Songs)-i-1].Name)
				player.Playlist[i].Name = v.Name
				player.Playlist[i].PlaySourceType = control.SourceWeb
				player.Playlist[i].ID = v.ID
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
			for i := range player.Playlist {
				fmt.Printf("\n[%d] %s", len(player.Playlist)-i, player.Playlist[len(player.Playlist)-i-1].Name)
			}
		}
		break
	case "ls":
		if len(player.Playlist) > 0 {
			for i := range player.Playlist {
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
		index--
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
		index--
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
		if player.PlayFeature != ModeNormal {
			logger.WriteLog("Please choose a play sheet first,Use: sheet [ID]")
		} else {
			player.PlayFeature = ModeIntel
			//log.Println(fmt.Sprintf(account.APIHeartbeatMode,player.Playlist[player.GetCurrentIndex()].ID,player.NowPlayingSheetID))
			reqURL := fmt.Sprintf(account.APIHeartbeatMode, player.Playlist[player.GetCurrentIndex()].ID, player.NowPlayingSheetID)
			listDatas := login.Client.DoGet(reqURL, nil)
			if listDatas == nil {
				fmt.Printf("\n[ERR] 请检查您的网络状况")
				break
			}
			var listDatasJSON interface{}
			var ss []account.Song
			json.Unmarshal(listDatas, &listDatasJSON)
			data := listDatasJSON.(map[string]interface{})["data"]
			if data != nil {
				for _, v := range data.([]interface{}) {
					var s account.Song
					songInfo := v.(map[string]interface{})["songInfo"]
					if songInfo != nil {
						ID := songInfo.(map[string]interface{})["ID"]
						name := songInfo.(map[string]interface{})["name"]
						//log.Println(name)
						if ID == nil || name == nil {
							continue
						}
						s.ID = int(ID.(float64))
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
				player.Playlist[i].ID = v.ID
				player.Playlist[i].PlaySourceType = control.SourceWeb
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
		Menu := Menu{Items: []string{}}
		fmt.Printf("\n===指令列表===")
		Menu.AddItem("[author] 显示作者信息")
		Menu.AddItem("[exit/q] 退出程序")
		Menu.AddItem("[login] <邮箱> <密码>: 邮箱登陆")
		Menu.AddItem("[logout]: 登出")
		Menu.AddItem("[qd]: 每日签到")
		Menu.AddItem("[fm]: 前往私人FM模式")
		Menu.AddItem("[fav]: 前往我喜欢的音乐")
		Menu.AddItem("[day]: 前往每日推荐")
		Menu.AddItem("[sheet]: 显示当前歌单列表")
		Menu.AddItem("[sheet] <序号>: 前往对应序号歌单")
		Menu.AddItem("[list/ls]: 显示播放列表")
		Menu.AddItem("[goto/go] <序号>: 转跳到指定序号歌曲")
		Menu.AddItem("[time/t] <sec>: 跳到歌曲的第sec秒")
		Menu.AddItem("[last/l]: 上一首")
		Menu.AddItem("[next/n]: 下一首")
		Menu.AddItem("[play/p]: 播放歌曲")
		Menu.AddItem("[pause]: 暂停歌曲")
		Menu.AddItem("[stop]: 停止歌曲")
		Menu.AddItem("[pg]: 显示进度条 #显示的时候输入字符会被刷掉.")
		Menu.AddItem("[key] 显示快捷键列表")
		c.PrintMenu(Menu)
		fmt.Printf("\n====================")
		break
	case "vol":
		value, err := strconv.Atoi(args[1])
		if value < 0 || value > 100 {
			fmt.Printf("\n[INFO] 错误的值(过大/过小)")
		}
		if err != nil {
			fmt.Printf("\n[INFO] 错误的值")
		}
		if player.SetVolume(uint(value)) == 1 {
			fmt.Printf("\n[INFO] 设置音量为 %d", value)
		} else {
			fmt.Printf("\n[INFO] 设置音量失败")
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
	case "q":
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

// PrintPrefix 打印每行的前缀
func (c *ClientApp) PrintPrefix(prefix string) {
	fmt.Printf("\n%s", prefix)
}

// PrintMenu 打印帮助菜单
func (c *ClientApp) PrintMenu(Menu Menu) {
	if !c.ShouldPrintMenu {
		return
	}
	for i, v := range Menu.Items {
		fmt.Printf("\n%d) %s", i, v)
	}
}

// AddItem 向帮助菜单中添加项目
func (m *Menu) AddItem(value string) *Menu {
	if m.Items != nil {
		m.Items = append(m.Items, value)
		return m
	}
	fmt.Printf("\n[INFO] 菜单项目不应为空")
	return nil
}

// ShowLyric 显示歌词
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

// GetLyricCurrent 获取当前播放歌曲的歌词
func GetLyricCurrent(player *control.Player, login *account.Login) string {
	ID := player.Playlist[player.GetCurrentIndex()].ID
	reqURL := fmt.Sprintf(account.APILyric, ID)
	lyricData := login.Client.DoGet(reqURL, nil)
	if lyricData == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况")
		return ""
	}
	var dataJSON interface{}
	json.Unmarshal(lyricData, &dataJSON)
	if dataJSON != nil {
		lrc := dataJSON.(map[string]interface{})["lrc"]
		if lrc != nil {
			lyric := lrc.(map[string]interface{})["lyric"]
			if lyric != nil {
				return lyric.(string)
			}
		}
	}
	return ""
}

// LyricLine 歌词行信息
type LyricLine struct {
	Min     int
	Sec     int
	Content string
	Length  int
}

// ParseLyric 解析歌词
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
