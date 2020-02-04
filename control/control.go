package control

import "C"
import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/CxZMoE/NetEase-CMD/account"
	"github.com/CxZMoE/NetEase-CMD/logger"
	bass "github.com/CxZMoE/bass-go"
	tm "github.com/buger/goterm"
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

//type handle string
//type File string

// Player 音乐播放器
type Player struct {
	Status            uint
	PlayFeature       int
	PlayMode          int
	LastSecond        int
	CurrentLength     int
	Playlist          []Music
	NowPlayingIndex   int
	LastIndex         int
	NowPlayingSheetID int
	LastHandle        uint
	Login             *account.Login
	LyricSwitch       bool
	Volume            uint
	IsShowProgress    bool
}

// Music 歌曲
type Music struct {
	ID             int
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
	SourceFile = 0
	SourceWeb  = 1
)

func init() {
	bass.Init()
	bass.PluginLoad("./lib/libbassflac.so")
}

// StartAPI 启动网易云API
func StartAPI() *exec.Cmd {
	homedir, err := os.UserHomeDir()
	if err != nil {
		logger.WriteLog("无法获取用户目录地质")
		return nil
	}
	apiExecPath := homedir + "/xzp/NeteaseApi/app.js"
	_, err = os.Stat(apiExecPath)
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

// NewPlayer 创建新播放器
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

	player.NowPlayingSheetID = sheet.ID // For heartbeat mode purpose
	for i, v := range sheet.Songs {
		player.Playlist[i].Name = v.Name
		player.Playlist[i].PlaySourceType = SourceWeb
		player.Playlist[i].ID = v.ID
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

// Release 释放所有播放资源，不能再播放除非重新加载。
func Release() {
	bass.Free()
}

// AttachFile 添加一个文件到播放器的播放列表
func (p *Player) AttachFile(file string) {
	music := Music{
		Name:           "",
		Author:         "",
		Album:          "",
		Cover:          "",
		Length:         0,
		FilePath:       file,
		Handle:         0,
		PlaySourceType: SourceFile,
	}
	p.Playlist = append(p.Playlist, music)
}

// AttachFileWeb 添加一个网络文件到播放器播放列表
func (p *Player) AttachFileWeb(url string) {
	music := Music{
		Name:           "",
		Author:         "",
		Album:          "",
		Cover:          "",
		Length:         0,
		FilePath:       url,
		Handle:         0,
		PlaySourceType: SourceWeb,
	}
	p.Playlist = append(p.Playlist, music)
}

// GetCurrentIndex 获取播放列表当前播放序号
func (p *Player) GetCurrentIndex() int {
	nowPlayingIndex := p.NowPlayingIndex
	return nowPlayingIndex
}

// SetCurrentIndex 设置播放列表当前播放序号
func (p *Player) SetCurrentIndex(index int) {
	p.NowPlayingIndex = index
}

// EmptyPlayList 清空播放列表
func (p *Player) EmptyPlayList() {
	p.Playlist = []Music{}
}

// GetNextIndex 获取下一首应该是那个序号
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

// GetLastIndex 获取下一首应该是那个序号
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

// Play 播放音乐
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
		if p.Playlist[nowPlayingIndex].PlaySourceType == SourceFile {
			handle = bass.StreamCreateFile(0, nowPlayingPath, 0, 0)
		}
		if p.Playlist[nowPlayingIndex].PlaySourceType == SourceWeb {
			p.RefreshPlayURL()
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

// Pause 暂停播放
func (p *Player) Pause() {
	handle := p.Playlist[p.GetCurrentIndex()].Handle
	if handle <= 0 {
		return
	}
	bass.ChannelPause(handle)
	p.LastSecond = p.Playlist[p.GetCurrentIndex()].GetPosition()
	p.Status = StatusPaused
}

// Stop 停止播放
func (p *Player) Stop() {
	handle := p.Playlist[p.GetCurrentIndex()].Handle
	if handle <= 0 {
		return
	}
	bass.ChannelStop(handle)
	p.Status = StatusStopped
}

// GetRandomIndex 获取随机的播放序号，用于随机模式。
func (p *Player) GetRandomIndex() int {
	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(p.Playlist))
	return randomIndex
}

// Next 下一首
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

// Last 上一首，如果是随机模式则为上一次播放的歌曲
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

// SetPlayMode 设置播放模式 [列表循环|单曲循环|随即|单曲结束等]
func (p *Player) SetPlayMode(mode int) {
	p.PlayMode = mode
}

// SetPlayFeature 设置播放特点 [我喜欢的|FM|日推等]
func (p *Player) SetPlayFeature(fea int) {
	p.PlayFeature = fea
}

// FreeStream 释放handle占用的资源
func (p *Player) FreeStream(handle uint) uint32 {
	return bass.StreamFree(handle)
}

// RemoveFile 从播放列表中移除制定序号歌曲
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

// GetPosition 获取播放位置（秒）
func (m *Music) GetPosition() int {
	return bass.ChannelBytes2Seconds(m.Handle, bass.ChannelGetPosition(m.Handle, bass.BASS_POS_BYTE))
}

// GetLength 获取歌曲长度（秒）
func (m *Music) GetLength() int {
	return bass.ChannelBytes2Seconds(m.Handle, bass.ChannelGetLength(m.Handle, bass.BASS_POS_BYTE))
}

// SetPosition 设置播放位置（秒）
func (m *Music) SetPosition(sec int) int {
	pos := bass.ChannelSeconds2Bytes(m.Handle, sec)
	return bass.ChannelSetPosition(m.Handle, pos, bass.BASS_POS_BYTE)
}

// SetCover 设置歌曲封面
func (m *Music) SetCover(src string) {
	m.Cover = src
}

// RefreshPlayURL 刷新歌曲的URL，建议每次播放前执行一次，防止链接实效。
func (p *Player) RefreshPlayURL() {
	defer func() {
		//捕获test抛出的panic
		if err := recover(); err != nil {
			fmt.Printf("\n[ERR] 请检查您的网络状况.")
			logger.WriteLog(fmt.Sprint(err))
		}
	}()
	url := p.Login.GetURLByID(p.Playlist[p.GetCurrentIndex()].ID)
	if url == "" {
		logger.WriteLog("Failed to refresh music url.")
		fmt.Printf("\n[ERR] 刷新歌曲播放地址失败.")
	}
	p.Playlist[p.GetCurrentIndex()].FilePath = url
}

// GetVolume 获取音量(0-100)%
func (p *Player) GetVolume() uint {
	return bass.GetChanVol(p.Playlist[p.GetCurrentIndex()].Handle)
}

// SetVolume 设置音量(0-100)%
func (p *Player) SetVolume(value uint) uint {
	p.Volume = value
	return bass.SetChanVol(p.Playlist[p.GetCurrentIndex()].Handle, value)
}
