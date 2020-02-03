package control

import "C"
import (
	"fmt"
	bass "github.com/CxZMoE/bass-go"
	"github.com/CxZMoE/xz-ease-player/account"
	"github.com/CxZMoE/xz-ease-player/logger"
	tm "github.com/buger/goterm"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// Play Status
const (
	NoStatus      = -1
	StatusPlaying = 0
	StatusPaused  = 1
	StatusStopped = 2
	StatusOther   = 3
)

// Play Modes
const (
	ModeListLoop   = 0
	ModeSingleLoop = 1
	ModeRandom     = 2
	ModeSingleStop = 3
)

type handle string
type File string

type Player struct {
	Status            uint
	PlayFeature       int
	PlayMode          int
	LastSecond        int
	CurrentLength     int
	Playlist          []Music
	NowPlayingIndex   int
	LastIndex         int
	NowPlayingSheetId int
	LastHandle        uint
	Login             *account.Login
	LyricSwitch       bool
	Volume            uint
	IsShowProgress    bool
}

type Music struct {
	Id             int
	Name           string
	Author         string
	Album          string
	Cover          string
	Length         int
	FilePath       string
	Handle         uint
	PlaySourceType int
}

// Play Source Type
const (
	SOURCE_FILE = 0
	SOURCE_WEB  = 1
)

func init() {
	bass.Init()
	bass.PluginLoad("./lib/libbassflac.so")
}

func StartAPI() *exec.Cmd {
	homedir, err := os.UserHomeDir()
	if err != nil {
		logger.WriteLog("无法获取用户目录地质")
		return nil
	}
	apiExecPath := homedir + "/xzp/NeteaseApi/app.js"
	_, err = os.Stat(apiExecPath)
	if os.IsNotExist(err) {
		logger.WriteLog("Couldn't start api server,app.js not found.")
		panic(err)
		return nil
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

func NewPlayer(login *account.Login) *Player {
	player := &Player{
		PlayMode:        ModeSingleLoop,
		Status:          StatusStopped,
		Playlist:        []Music{},
		NowPlayingIndex: 0,
		Login:           login,
		LyricSwitch:     false,
		Volume:          100,
		LastIndex:       0,
		IsShowProgress:  false,
	}
	login.GetAllPlaySheet()
	if len(login.PlaySheets) == 0 {
		fmt.Printf("[ERR] 获取我喜欢的歌单失败.\n[ERR]请用[fav]指令重试.")
		return player
	}
	sheet := login.PlaySheets[0]
	// Load Detail
	sheet.LoadDetail(login) // Load Song Ids.
	player.PlayFeature = 2  // MODE_NORMAL
	player.EmptyPlayList()

	player.Playlist = make([]Music, len(sheet.Songs))

	player.NowPlayingSheetId = sheet.Id // For heartbeat mode purpose
	for i, v := range sheet.Songs {
		player.Playlist[i].Name = v.Name
		player.Playlist[i].PlaySourceType = SOURCE_WEB
		player.Playlist[i].Id = v.Id
	}
	player.SetCurrentIndex(0)
	player.PlayMode = StatusOther

	go func() {
		for {
			if player.IsShowProgress == true {
				pos := player.Playlist[player.GetCurrentIndex()].GetPosition()
				length := player.Playlist[player.GetCurrentIndex()].GetLength()

				if length > 0 {
					tm.MoveCursor(tm.Width()-10, tm.Height()-1)
					tm.Printf("[%d/%d]", pos, length)
					tm.MoveCursor(4, tm.Height())
					tm.Flush()
				}
			}
			time.Sleep(time.Millisecond * 100)
		}

	}()
	return player
}

func Release() {
	bass.Free()
}

func (p *Player) AttachFile(file string) {
	music := Music{
		Name:           "",
		Author:         "",
		Album:          "",
		Cover:          "",
		Length:         0,
		FilePath:       file,
		Handle:         0,
		PlaySourceType: SOURCE_FILE,
	}
	p.Playlist = append(p.Playlist, music)
}

func (p *Player) AttachFileWeb(url string) {
	music := Music{
		Name:           "",
		Author:         "",
		Album:          "",
		Cover:          "",
		Length:         0,
		FilePath:       url,
		Handle:         0,
		PlaySourceType: SOURCE_WEB,
	}
	p.Playlist = append(p.Playlist, music)
}

func (p *Player) GetCurrentIndex() int {
	nowPlayingIndex := p.NowPlayingIndex
	return nowPlayingIndex
}

func (p *Player) SetCurrentIndex(index int) {
	p.NowPlayingIndex = index
}

func (p *Player) EmptyPlayList() {
	p.Playlist = []Music{}
}

func (p *Player) GetNextIndex() int {
	listLen := len(p.Playlist)
	currentIndex := p.GetCurrentIndex()
	if listLen <= 0 {
		return 0
	}
	if currentIndex == (len(p.Playlist) - 1) {
		currentIndex = 0
	} else {
		currentIndex = currentIndex + 1
	}
	return currentIndex
}

func (p *Player) GetLastIndex() int {
	listLen := len(p.Playlist)
	currentIndex := p.GetCurrentIndex()

	if listLen <= 0 {
		return 0
	}
	if currentIndex == 0 {
		currentIndex = listLen - 1
	} else {
		currentIndex = currentIndex - 1
	}
	return currentIndex
}
func (p *Player) Play() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("\n[ERR] 请检查您的网络状况.")
			logger.WriteLog(fmt.Sprint(err))
		}
	}()
	if len(p.Playlist) <= 0 {
		fmt.Print("\n[ERR] 播放列表为空.")
		return
	}
	// When Finished Playing
	restart := 0

	if p.PlayMode == ModeSingleLoop {
		restart = 1
	}
	if p.PlayMode == ModeSingleStop {
		restart = 0
	}

	var handle uint

	//Play Status
	if p.Status == StatusPlaying || p.Status == StatusOther || p.Status == StatusStopped {
		nowPlayingIndex := p.GetCurrentIndex()
		nowPlayingPath := p.Playlist[nowPlayingIndex].FilePath
		if p.Playlist[nowPlayingIndex].PlaySourceType == SOURCE_FILE {
			handle = bass.StreamCreateFile(0, nowPlayingPath, 0, 0)
		}
		if p.Playlist[nowPlayingIndex].PlaySourceType == SOURCE_WEB {
			p.RefreshPlayUrl()
			nowPlayingPath := p.Playlist[nowPlayingIndex].FilePath
			//log.Println("Now",nowPlayingPath)
			handle = bass.StreamCreateURL(nowPlayingPath, 0, nil, nil)
		}
		p.Stop()
		p.FreeStream(p.LastHandle)
		p.LastHandle = handle
		p.Playlist[nowPlayingIndex].Handle = handle
		handle = p.Playlist[nowPlayingIndex].Handle
		p.SetVolume(p.Volume)
		fmt.Printf("\n[VOLUME] %d%%", int(p.GetVolume()))
		fmt.Printf("\n[INDEX] %d", p.GetCurrentIndex()+1)
		fmt.Printf("\n[INFO] 正在播放: %s", p.Playlist[nowPlayingIndex].Name)
		parts := strings.Split(p.Playlist[nowPlayingIndex].FilePath, ".")
		ext := parts[len(parts)-1]
		if ext == "" {
			p.Next()
		}
		fmt.Printf("\n[INFO] 格式: %s", ext)

		p.CurrentLength = p.Playlist[p.GetCurrentIndex()].GetLength()
		fmt.Printf("\n[INFO] 长度: %d秒", p.CurrentLength)
		bass.ChannelPlay(handle, restart)
		p.Status = StatusPlaying
	}
	if p.Status == StatusPaused {
		handle = p.Playlist[p.GetCurrentIndex()].Handle
		bass.ChannelPlay(handle, restart)
		p.Playlist[p.GetCurrentIndex()].SetPosition(p.LastSecond)
		p.Status = StatusPlaying
	}
}

func (p *Player) Pause() {
	handle := p.Playlist[p.GetCurrentIndex()].Handle
	if handle <= 0 {
		return
	}
	bass.ChannelPause(handle)
	p.LastSecond = p.Playlist[p.GetCurrentIndex()].GetPosition()
	p.Status = StatusPaused
}

func (p *Player) Stop() {
	handle := p.Playlist[p.GetCurrentIndex()].Handle
	if handle <= 0 {
		return
	}
	bass.ChannelStop(handle)
	p.Status = StatusStopped
}

func (p *Player) GetRandomIndex() int {
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(p.Playlist))
	return randomIndex
}
func (p *Player) Next() {
	defer func() {
		//捕获test抛出的panic
		if err := recover(); err != nil {
			fmt.Printf("\n[ERR] 请检查您的网络状况.")
			logger.WriteLog(fmt.Sprint(err))
		}
	}()
	listLen := len(p.Playlist)
	currentIndex := p.GetCurrentIndex()
	if listLen <= 0 {
		logger.WriteLog("播放列表为空")
		return
	}
	if p.PlayMode == ModeRandom {
		p.SetCurrentIndex(p.GetRandomIndex())
	} else {
		if currentIndex == (len(p.Playlist) - 1) {
			p.SetCurrentIndex(0)
		} else {
			p.SetCurrentIndex(currentIndex + 1)
		}
	}
	bass.StreamFree(p.Playlist[currentIndex].Handle)
	if p.Status == StatusPaused {

		p.Status = StatusPlaying
	}

	p.Play()
}

func (p *Player) PrintLog() {
	fmt.Printf("\n[INFO] A Test Msg...")
}
func (p *Player) Last() {
	defer func() {
		//捕获test抛出的panic
		if err := recover(); err != nil {
			fmt.Printf("\n[ERR] 请检查您的网络状况.")
			logger.WriteLog(fmt.Sprint(err))
		}
	}()
	listLen := len(p.Playlist)
	currentIndex := p.GetCurrentIndex()
	if listLen <= 0 {
		return
	}

	if p.PlayMode == ModeRandom {
		p.SetCurrentIndex(p.LastIndex)
	} else {
		if currentIndex == 0 {
			p.SetCurrentIndex(listLen - 1)
		} else {
			p.SetCurrentIndex(currentIndex - 1)
		}

	}

	bass.StreamFree(p.Playlist[currentIndex].Handle)

	p.Play()
}

func (p *Player) SetPlayMode(mode int) {
	p.PlayMode = mode
}

func (p *Player) SetPlayFeature(fea int) {
	p.PlayFeature = fea
}

func (p *Player) FreeStream(handle uint) uint32 {
	return bass.StreamFree(handle)
}

func (p *Player) RemoveFile(index int) {
	if len(p.Playlist) <= 0 {
		return
	}
	if p.Playlist[index].Handle != 0 {
		p.Playlist = append(p.Playlist[:index], p.Playlist[index+1:]...)
	}
	if p.GetCurrentIndex() > index {
		p.SetCurrentIndex(p.GetCurrentIndex() - 1)
	}
}

func (m *Music) GetPosition() int {
	return bass.ChannelBytes2Seconds(m.Handle, bass.ChannelGetPosition(m.Handle, bass.BASS_POS_BYTE))
}

func (m *Music) GetLength() int {
	return bass.ChannelBytes2Seconds(m.Handle, bass.ChannelGetLength(m.Handle, bass.BASS_POS_BYTE))
}

func (m *Music) SetPosition(sec int) int {
	pos := bass.ChannelSeconds2Bytes(m.Handle, sec)
	return bass.ChannelSetPosition(m.Handle, pos, bass.BASS_POS_BYTE)
}

func (m *Music) SetCover(src string) {
	m.Cover = src
}

func (p *Player) RefreshPlayUrl() {
	defer func() {
		//捕获test抛出的panic
		if err := recover(); err != nil {
			fmt.Printf("\n[ERR] 请检查您的网络状况.")
			logger.WriteLog(fmt.Sprint(err))
		}
	}()
	url := p.Login.GetUrlById(p.Playlist[p.GetCurrentIndex()].Id)
	if url == "" {
		logger.WriteLog("Failed to refresh music url.")
		fmt.Printf("\n[ERR] 刷新歌曲播放地址失败.")
	}
	p.Playlist[p.GetCurrentIndex()].FilePath = url
}

func (p *Player) GetVolume() uint {
	return bass.GetChanVol(p.Playlist[p.GetCurrentIndex()].Handle)
}

func (p *Player) SetVolume(value uint) uint {
	p.Volume = value
	return bass.SetChanVol(p.Playlist[p.GetCurrentIndex()].Handle, value)
}
