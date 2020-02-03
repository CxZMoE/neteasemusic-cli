package account

import (
	"encoding/json"
	"fmt"
	"github.com/CxZMoE/xz-ease-player/logger"
	"github.com/CxZMoE/xz-ease-player/network"
	"log"
)

const (
	ApiServer             = "http://localhost:3000/"
	ApiLoginMail          = ApiServer + "login?email=%s&password=%s"
	ApiLoginStatus        = ApiServer + "login/status"
	ApiLogout             = ApiServer + "logout"
	ApiUserDetail         = ApiServer + "user/detail?uid=%d"
	ApiHeartbeatMode      = ApiServer + "playmode/intelligence/list?id=%d&pid=%d" // songid and playsheerid
	ApiMusicUrl           = ApiServer + "song/url?id="                            // More can be added
	ApiMusicUrlSingle     = ApiServer + "song/url?id=%d"
	ApiSignIn             = ApiServer + "daily_signin?type=%d"
	ApiFM                 = ApiServer + "personal_fm"
	ApiDailyRecommendSong = ApiServer + "recommend/songs"
	ApiLoveList           = ApiServer + "likelist?uid=%d"
	ApiUserPlaySheetList  = ApiServer + "user/playlist?uid=%d"
	ApiPlaySheetDetail    = ApiServer + "playlist/detail?id=%d"
	ApiSongDetail         = ApiServer + "song/detail?ids="
	ApiLyric              = ApiServer + "lyric?id=%d"
	ApiLike               = ApiServer + "like?id=%d"
)

type Login struct {
	ApiAddr    string
	Client     *network.Client
	LoginData  []byte
	UserData   User
	PlaySheets []PlaySheet
}

type User struct {
	Uid        int
	UserName   string
	CreateTime int
	VipType    int
	Profile    Profile
}

type Profile struct {
	NickName      string
	Gender        int
	Birthday      int
	DefaultAvatar bool
	AvatarUrl     string
	City          int
}

func (l *Login) Like(id int) bool {
	reqUrl := fmt.Sprintf(ApiLike, id)
	res := l.Client.DoGet(reqUrl, nil)
	if res == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况")
		return false
	}
	if l.GetStatusCode(res) == 200 {
		return true
	}
	return false
}

func (l *Login) GetStatusCode(data []byte) int {
	var j interface{}
	err := json.Unmarshal(data, &j)
	if err != nil {
		logger.WriteLog(err.Error())
	}
	if j != nil {
		code := int(j.(map[string]interface{})["code"].(float64))
		return code
	} else {
		return 0
	}

}

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
		l.UserData.Uid = int(account["id"].(float64))
		l.UserData.UserName = string(account["userName"].(string))
		l.UserData.CreateTime = int(account["createTime"].(float64))
		l.UserData.VipType = int(account["vipType"].(float64))

		// Node Profile:
		profile := j.(map[string]interface{})["profile"].(map[string]interface{})
		l.UserData.Profile.NickName = profile["nickname"].(string)
		l.UserData.Profile.AvatarUrl = profile["avatarUrl"].(string)
		l.UserData.Profile.Birthday = int(profile["birthday"].(float64))
		l.UserData.Profile.City = int(profile["city"].(float64))
		l.UserData.Profile.DefaultAvatar = false
		l.UserData.Profile.Gender = int(profile["gender"].(float64))
	} else {
		return
	}

}

func (l *Login) LoginEmail(email, password string) []byte {
	url := fmt.Sprintf(ApiLoginMail, email, password)
	data := l.Client.DoLoginGet(url, nil)
	l.LoginData = data
	//log.Println(string(data))
	l.ParseLoginData()
	l.Client.SaveJar(l.Client.CoreClient.Jar)
	return data
}

func (l *Login) GetUid() int {
	var j map[string]interface{}
	json.Unmarshal(l.LoginData, &j)
	if j != nil {
		account := j["account"].(map[string]interface{})
		return int(account["id"].(float64))
	} else {
		return 0
	}
}

func (l *Login) Logout() []byte {
	result := l.Client.DoGet(ApiLogout, nil)
	if result == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况")
		return nil
	}
	return result
}

func (l *Login) GetUrlById(id int) string {
	var j interface{}
	data := l.Client.DoGet(fmt.Sprintf(ApiMusicUrlSingle, id), nil)
	if data == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况")
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

type Song struct {
	Name string
	Url  string
	Id   int
}

func (l *Login) GetFMSong() []Song {
	var j interface{}
	var ss []Song
	data := l.Client.DoGet(ApiFM, nil)
	if data == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况")
		return nil
	}
	json.Unmarshal(data, &j)
	if data != nil {
		list := j.(map[string]interface{})["data"].([]interface{})
		for i, _ := range list {
			var s Song
			s.Name = list[i].(map[string]interface{})["name"].(string)
			s.Url = l.GetUrlById(int(list[i].(map[string]interface{})["id"].(float64)))
			if s.Url == "" {
				continue
			}
			s.Id = int(list[i].(map[string]interface{})["id"].(float64))
			ss = append(ss, s)
		}
	} else {
		return nil
	}

	return ss
}

func (l *Login) GetFavSong() []Song {
	var j interface{}
	var ss []Song
	data := l.Client.DoGet(fmt.Sprintf(ApiLoveList, l.UserData.Uid), nil)
	if data == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况")
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

type PlaySheet struct {
	Id          int
	Name        string
	Creator     int
	CreatorName string
	Songs       []Song
}

// Load detail(id,name) of a play sheet.
func (p *PlaySheet) LoadDetail(l *Login) {
	var ss []Song
	reqUrl := fmt.Sprintf(ApiPlaySheetDetail, p.Id)
	detailDatas := l.Client.DoGet(reqUrl, nil)
	if detailDatas == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况")
		return
	}
	var detailDatasJson interface{}
	json.Unmarshal(detailDatas, &detailDatasJson)
	if detailDatasJson != nil {
		playList := detailDatasJson.(map[string]interface{})["playlist"]
		if playList != nil {
			tracks := playList.(map[string]interface{})["tracks"]
			if tracks != nil {
				for _, v := range tracks.([]interface{}) {
					var s Song
					songName := v.(map[string]interface{})["name"]
					songId := v.(map[string]interface{})["id"]
					if songName == nil || songId == nil {
						continue
					} else {
						s.Name = songName.(string)
						s.Id = int(songId.(float64))
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

func (l *Login) GetAllPlaySheet() {
	uid := l.UserData.Uid
	reqUrl := fmt.Sprintf(ApiUserPlaySheetList, uid)
	listDatas := l.Client.DoGet(reqUrl, nil)
	if listDatas == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况")
		return
	}
	var playListJson interface{}
	var sheets []PlaySheet
	json.Unmarshal(listDatas, &playListJson)
	if playListJson != nil {
		playList := playListJson.(map[string]interface{})["playlist"]
		if playList != nil {
			for _, v := range playList.([]interface{}) {
				var s PlaySheet
				sheetId := v.(map[string]interface{})["id"]
				sheetName := v.(map[string]interface{})["name"]
				creator := v.(map[string]interface{})["creator"]
				if sheetId != nil {
					s.Id = int(sheetId.(float64))
				} else {
					continue
				}
				if sheetName != nil {
					s.Name = sheetName.(string)
				} else {
					continue
				}
				if creator != nil {
					creatorId := creator.(map[string]interface{})["userId"]
					if creatorId != nil {
						s.Creator = int(creatorId.(float64))
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
		l.PlaySheets = sheets
	} else {
		return
	}
}

func (l *Login) GetSongDetails(ids ...int) []Song {
	var songs_temp []Song
	var nameUrl = ApiSongDetail
	for i, v := range ids {
		var s Song
		s.Id = v
		songs_temp = append(songs_temp, s)
		if i == len(ids)-1 {
			nameUrl += fmt.Sprintf("%d", v)
			continue
		}
		nameUrl += fmt.Sprintf("%d,", v)
	}

	//log.Println("Created Url...")
	var songs []Song
	//log.Println("GetSongDetails...")
	songDatas := l.Client.DoGet(nameUrl, nil)
	if songDatas == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况")
		return nil
	}
	//log.Println("GetSongDetails...")
	var songData interface{}

	if err := json.Unmarshal(songDatas, &songData); err != nil {
		log.Println(err)
	}
	// song/url
	// song/url

	if songData != nil {
		songsJson := songData.(map[string]interface{})["songs"].([]interface{})
		// song/url
		for _, v := range songsJson {
			var s Song
			id := v.(map[string]interface{})["id"]
			if id == nil {
				continue
			} else {
				name := v.(map[string]interface{})["name"]
				if name == nil {
					continue
				}
				s.Id = int(id.(float64))
				s.Name = name.(string)
				// Make sure we both have id and name
			}
			songs = append(songs, s)

		}
	} else {
		return nil
	}

	return songs
}

func (l *Login) GetRecommend() []Song {
	var recJson interface{}
	var ss []Song
	reqUrl := ApiDailyRecommendSong
	recDatas := l.Client.DoGet(reqUrl, nil)
	if recDatas == nil {
		fmt.Printf("\n[ERR] 请检查您的网络状况")
		return nil
	}
	json.Unmarshal(recDatas, &recJson)
	if recDatas != nil {
		recommend := recJson.(map[string]interface{})["recommend"]
		if recommend != nil {
			for _, v := range recommend.([]interface{}) {
				var s Song
				if v != nil {
					name := v.(map[string]interface{})["name"]
					id := v.(map[string]interface{})["id"]
					if name == nil || id == nil {
						continue
					} else {
						s.Name = name.(string)
						s.Id = int(id.(float64))
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

func NewLogin(client *network.Client) *Login {

	return &Login{"http://localhost:3000/", client, nil, User{}, []PlaySheet{}}
}
