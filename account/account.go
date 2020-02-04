package account

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/CxZMoE/NetEase-CMD/logger"
	"github.com/CxZMoE/NetEase-CMD/network"
)

// API 地址列表
const (
	// APIServer API服务器地址
	APIServer = "http://localhost:3000/"
	// APILoginMail 邮箱登录地址
	APILoginMail = APIServer + "login?email=%s&password=%s"
	// APILoginStatus 登录状态地址
	APILoginStatus = APIServer + "login/status"
	// APILogout 登出
	APILogout = APIServer + "logout"
	// APIUserDetail 用户详情
	APIUserDetail = APIServer + "user/detail?uid=%d"
	// APIHeartbeatMode 心动模式
	APIHeartbeatMode = APIServer + "playmode/intelligence/list?id=%d&pid=%d" // songID and playsheerid
	// APIMusicURL 获取音乐的地址
	APIMusicURL = APIServer + "song/url?id=" // More can be added
	// APIMusicURLSingle 获取单个音乐地址
	APIMusicURLSingle = APIServer + "song/url?id=%d"
	// APISignIn 签到API
	APISignIn = APIServer + "daily_signin?type=%d"
	// APIFM 电台API
	APIFM = APIServer + "personal_fm"
	// APIDailyRecommendSong 每日推荐API
	APIDailyRecommendSong = APIServer + "recommend/songs"
	// APILoveList 我喜欢的音乐API
	APILoveList = APIServer + "likelist?uid=%d"
	// APIUserPlaySheetList 用户歌单API
	APIUserPlaySheetList = APIServer + "user/playlist?uid=%d"
	// APIPlaySheetDetail 歌单详情API
	APIPlaySheetDetail = APIServer + "playlist/detail?id=%d"
	// APISongDetail 歌曲详情API
	APISongDetail = APIServer + "song/detail?ids="
	// APILyric 歌词API
	APILyric = APIServer + "lyric?id=%d"
	// APILike 喜欢此音乐API
	APILike = APIServer + "like?id=%d"
)

// Login 网易云登录客户端
type Login struct {
	APIAddr    string
	Client     *network.Client
	LoginData  []byte
	UserData   User
	PlaySheets []PlaySheet
}

// User 用户信息登记
type User struct {
	UID        int
	UserName   string
	CreateTime int
	VipType    int
	Profile    Profile
}

// Profile 用户个人信息登记
type Profile struct {
	NickName      string
	Gender        int
	Birthday      int
	DefaultAvatar bool
	AvatarURL     string
	City          int
}

// Like 喜欢指定ID的音乐
func (l *Login) Like(ID int) bool {
	reqURL := fmt.Sprintf(APILike, ID)
	res := l.Client.DoGet(reqURL, nil)
	if res == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况\n")
		return false
	}
	if l.GetStatusCode(res) == 200 {
		return true
	}
	return false
}

// DisLike 取消喜欢指定ID的音乐
func (l *Login) DisLike(ID int) bool {
	reqURL := fmt.Sprintf(APILike, ID) + "&like=false"
	res := l.Client.DoGet(reqURL, nil)
	if res == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况\n")
		return false
	}
	if l.GetStatusCode(res) == 200 {
		return true
	}
	return false
}

// GetStatusCode 获取返回状态码
func (l *Login) GetStatusCode(data []byte) int {
	var j interface{}
	err := json.Unmarshal(data, &j)
	if err != nil {
		logger.WriteLog(err.Error())
	}
	if j != nil {
		code := int(j.(map[string]interface{})["code"].(float64))
		return code
	}
	return 0
}

// ParseLoginData 解析登录信息
func (l *Login) ParseLoginData() {
	data := l.LoginData
	if data == nil {
		logger.WriteLog("LoginData is empty.")
		return
	}

	var j interface{}
	err := json.Unmarshal(data, &j)

	if err != nil {
		logger.WriteLog(err.Error())
	}
	if j != nil {
		// Node Account:
		account := j.(map[string]interface{})["account"].(map[string]interface{})
		l.UserData.UID = int(account["id"].(float64))
		l.UserData.UserName = string(account["userName"].(string))
		l.UserData.CreateTime = int(account["createTime"].(float64))
		l.UserData.VipType = int(account["vipType"].(float64))

		// Node Profile:
		profile := j.(map[string]interface{})["profile"].(map[string]interface{})
		l.UserData.Profile.NickName = profile["nickname"].(string)
		l.UserData.Profile.AvatarURL = profile["avatarurl"].(string)
		l.UserData.Profile.Birthday = int(profile["birthday"].(float64))
		l.UserData.Profile.City = int(profile["city"].(float64))
		l.UserData.Profile.DefaultAvatar = false
		l.UserData.Profile.Gender = int(profile["gender"].(float64))
	} else {
		return
	}

}

// LoginEmail 使用邮箱登录
func (l *Login) LoginEmail(email, password string) []byte {
	URL := fmt.Sprintf(APILoginMail, email, password)
	data := l.Client.DoLoginGet(URL, nil)
	l.LoginData = data
	//log.Println(string(data))
	l.ParseLoginData()
	l.Client.SaveJar(l.Client.CoreClient.Jar)
	return data
}

// GetUID 获取用户UID
func (l *Login) GetUID() int {
	var j map[string]interface{}
	json.Unmarshal(l.LoginData, &j)
	if j != nil {
		account := j["account"].(map[string]interface{})
		return int(account["id"].(float64))
	}
	return 0
}

// Logout 登出
func (l *Login) Logout() []byte {
	result := l.Client.DoGet(APILogout, nil)
	if result == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况\n")
		return nil
	}
	return result
}

// GetURLByID 获取歌曲URL
func (l *Login) GetURLByID(ID int) string {
	var j interface{}
	data := l.Client.DoGet(fmt.Sprintf(APIMusicURLSingle, ID), nil)
	if data == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况\n")
		return ""
	}
	json.Unmarshal(data, &j)
	var a interface{}
	if data != nil {
		a = j.(map[string]interface{})["data"].([]interface{})[0].(map[string]interface{})["url"]
		if a == nil {
			return ""
		}

	} else {
		return ""
	}
	return a.(string)
}

// Song 记录歌曲播放所需的信息，用于传递给播放列表信息。
type Song struct {
	Name string
	URL  string
	ID   int
}

// GetFMSong 获取私人电台歌曲列表
func (l *Login) GetFMSong() []Song {
	var j interface{}
	var ss []Song
	data := l.Client.DoGet(APIFM, nil)
	if data == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况\n")
		return nil
	}
	json.Unmarshal(data, &j)
	if data != nil {
		list := j.(map[string]interface{})["data"].([]interface{})
		for i := range list {
			var s Song
			s.Name = list[i].(map[string]interface{})["name"].(string)
			s.URL = l.GetURLByID(int(list[i].(map[string]interface{})["id"].(float64)))
			if s.URL == "" {
				continue
			}
			s.ID = int(list[i].(map[string]interface{})["id"].(float64))
			ss = append(ss, s)
		}
	} else {
		return nil
	}

	return ss
}

// GetFavSong 获取我喜欢的音乐歌曲列表
func (l *Login) GetFavSong() []Song {
	var j interface{}
	var ss []Song
	data := l.Client.DoGet(fmt.Sprintf(APILoveList, l.UserData.UID), nil)
	if data == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况\n")
		return nil
	}
	json.Unmarshal(data, &j)
	ids := j.(map[string]interface{})["ids"]
	if ids != nil {
		var idsInt []int
		for _, v := range ids.([]interface{}) {
			if v == nil {
				continue
			}
			idsInt = append(idsInt, int(v.(float64)))
		}
		//log.Println("GetDetails...")

		ss = l.GetSongDetails(idsInt...)

		//log.Println("GetDetails...")
	} else {
		return nil
	}

	return ss
}

// PlaySheet 用户歌单
type PlaySheet struct {
	ID          int
	Name        string
	Creator     int
	CreatorName string
	Songs       []Song
}

// LoadDetail 加载歌单的详情
func (p *PlaySheet) LoadDetail(l *Login) {
	var ss []Song
	reqURL := fmt.Sprintf(APIPlaySheetDetail, p.ID)
	detailDatas := l.Client.DoGet(reqURL, nil)
	if detailDatas == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况\n")
		return
	}
	var detailDatasJSON interface{}
	json.Unmarshal(detailDatas, &detailDatasJSON)
	if detailDatasJSON != nil {
		playList := detailDatasJSON.(map[string]interface{})["playlist"]
		if playList != nil {
			tracks := playList.(map[string]interface{})["tracks"]
			if tracks != nil {
				for _, v := range tracks.([]interface{}) {
					var s Song
					songName := v.(map[string]interface{})["name"]
					songID := v.(map[string]interface{})["id"]
					if songName == nil || songID == nil {
						continue
					} else {
						s.Name = songName.(string)
						s.ID = int(songID.(float64))
					}
					ss = append(ss, s)
				}
			} else {
				return
			}
		} else {
			return
		}
		p.Songs = ss
	} else {
		return
	}

}

// GetAllPlaySheet 获取用户所有的歌单
func (l *Login) GetAllPlaySheet() {
	UID := l.UserData.UID
	reqURL := fmt.Sprintf(APIUserPlaySheetList, UID)
	listDatas := l.Client.DoGet(reqURL, nil)
	if listDatas == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况\n\n")
		return
	}
	var playListJSON interface{}
	var sheets []PlaySheet
	json.Unmarshal(listDatas, &playListJSON)
	if playListJSON != nil {
		playList := playListJSON.(map[string]interface{})["playlist"]
		if playList != nil {
			for _, v := range playList.([]interface{}) {
				var s PlaySheet
				sheetID := v.(map[string]interface{})["id"]
				sheetName := v.(map[string]interface{})["name"]
				creator := v.(map[string]interface{})["creator"]
				if sheetID != nil {
					s.ID = int(sheetID.(float64))
				} else {
					continue
				}
				if sheetName != nil {
					s.Name = sheetName.(string)
				} else {
					continue
				}
				if creator != nil {
					creatorID := creator.(map[string]interface{})["userId"]
					if creatorID != nil {
						s.Creator = int(creatorID.(float64))
					} else {
						continue
					}
					creatorName := creator.(map[string]interface{})["nickname"]
					if creatorName != nil {
						s.CreatorName = creatorName.(string)
					} else {
						continue
					}
				}
				sheets = append(sheets, s)
			}
		} else {
			return
		}
		// 歌单赋值给login客户端
		l.PlaySheets = sheets
	} else {
		return
	}
}

// GetSongDetails 获取id序列的详细信息，包含URL
func (l *Login) GetSongDetails(ids ...int) []Song {
	var songsTemp []Song
	var nameURL = APISongDetail
	for i, v := range ids {
		var s Song
		s.ID = v
		songsTemp = append(songsTemp, s)
		if i == len(ids)-1 {
			nameURL += fmt.Sprintf("%d", v)
			continue
		}
		nameURL += fmt.Sprintf("%d,", v)
	}

	//log.Println("Created URL...")
	var songs []Song
	//log.Println("GetSongDetails...")
	songDatas := l.Client.DoGet(nameURL, nil)
	if songDatas == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况\n")
		return nil
	}
	//log.Println("GetSongDetails...")
	var songData interface{}

	if err := json.Unmarshal(songDatas, &songData); err != nil {
		log.Println(err)
	}
	// song/URL
	// song/URL

	if songData != nil {
		songsJSON := songData.(map[string]interface{})["songs"].([]interface{})
		// song/URL
		for _, v := range songsJSON {
			var s Song
			ID := v.(map[string]interface{})["id"]
			if ID == nil {
				continue
			} else {
				name := v.(map[string]interface{})["name"]
				if name == nil {
					continue
				}
				s.ID = int(ID.(float64))
				s.Name = name.(string)
				// Make sure we both have ID and name
			}
			songs = append(songs, s)

		}
	} else {
		return nil
	}

	return songs
}

// GetRecommend 获取每日推荐歌单
func (l *Login) GetRecommend() []Song {
	var recJSON interface{}
	var ss []Song
	reqURL := APIDailyRecommendSong
	recDatas := l.Client.DoGet(reqURL, nil)
	if recDatas == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况\n")
		return nil
	}
	json.Unmarshal(recDatas, &recJSON)
	if recDatas != nil {
		recommend := recJSON.(map[string]interface{})["recommend"]
		if recommend != nil {
			for _, v := range recommend.([]interface{}) {
				var s Song
				if v != nil {
					name := v.(map[string]interface{})["name"]
					ID := v.(map[string]interface{})["id"]
					if name == nil || ID == nil {
						continue
					} else {
						s.Name = name.(string)
						s.ID = int(ID.(float64))
					}
				} else {
					continue
				}
				ss = append(ss, s)
			}
		} else {
			return nil
		}
	} else {
		return nil
	}
	return ss
}

// NewLogin 创建新的网易云登录客户端
func NewLogin(client *network.Client) *Login {

	return &Login{"http://localhost:3000/", client, nil, User{}, []PlaySheet{}}
}
