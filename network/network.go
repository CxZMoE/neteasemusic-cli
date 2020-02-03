package network

import (
	"bytes"
	"fmt"
	"github.com/CxZMoE/xz-ease-player/logger"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	url2 "net/url"
	"os"
	"strings"
	"time"
)

type Headers map[string]string
type Params map[string]string
type PostForm map[string]string

// Client xz-ease-player client
type Client struct {
	LoginStatus bool
	CoreClient  http.Client
	UserAgent   string
}

// NewClient Create a new client for request.
func NewClient() *Client {
	homedir, err := os.UserHomeDir()
	if err != nil {
		logger.WriteLog("Failed to get home path.")
		return nil
	}
	_, err = os.Stat(homedir + "/xzp/cookie")
	var cookieJar http.CookieJar
	var loginStatus bool
	if os.IsExist(err) {
		cookieJar = LoadJar(homedir + "/xzp/cookie")
		loginStatus = true
	} else {
		cookieJar, err = cookiejar.New(nil)
		loginStatus = false
	}
	dealErr(err)

	httpClient := http.Client{
		Timeout: time.Second * 10,
		Jar:     cookieJar,
	}

	client := &Client{
		LoginStatus: loginStatus,
		CoreClient:  httpClient,
		UserAgent:   "xz-ease-player/1.0",
	}
	return client
}

// NewRequest Create new request
func (c *Client) NewRequest(method string, headers Headers, url string, params Params, body []byte) *http.Request {
	url = makeUrl(url, params)
	// Make Request
	request, err := http.NewRequest(strings.ToUpper(method), url, bytes.NewBuffer(body))
	dealErr(err)
	for k, v := range headers {
		request.Header.Add(k, v)
	}

	return request
}

func (c *Client) DoLoginGet(url string, params map[string]string) []byte {
	client := c.CoreClient

	// Create headers
	headers := Headers{
		"User-Agent": "xz-ease-player/1.0",
		"xhrFields":  "{ withCredentials: true }",
	}
	req := c.NewRequest("GET", headers, url, params, nil)
	// Receive Response
CLIENTDO1:
	resp, err := client.Do(req)
	if dealErr(err) != nil {
		logger.WriteLog("retrying...")
		time.Sleep(1 * time.Second)
		goto CLIENTDO1
	}
	defer resp.Body.Close()
	//u,_ := url2.Parse("http://localhost:3000/")
	//c.CoreClient.Jar.SetCookies(u,resp.Cookies())
	body, err := ioutil.ReadAll(resp.Body)
	dealErr(err)
	return body
}

// DoGet Do get request returns body
func (c *Client) DoGet(url string, params map[string]string) []byte {
	client := c.CoreClient

	// Create headers
	headers := Headers{
		"User-Agent": "xz-ease-player/1.0",
		"xhrFields":  "{ withCredentials: true }",
	}
	req := c.NewRequest("GET", headers, url, params, nil)
	//c.CoreClient.Jar = LoadJar("./cookie")
	// Receive Response
CLIENTDO1:
	resp, err := client.Do(req)
	if dealErr(err) != nil {
		logger.WriteLog("retrying..")
		time.Sleep(1 * time.Second)
		goto CLIENTDO1
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	dealErr(err)
	//log.Println(string(body))
	return body
}

// DoPostForm Do post request returns body
func (c *Client) DoPostForm(url string, req *http.Request, params Params, form PostForm, body []byte) []byte {
	client := c.CoreClient

	req.ParseForm()
	for k, v := range form {
		req.PostForm.Add(k, v)
	}
	u, _ := url2.Parse("http://localhost:3000/")
	for _, v := range c.CoreClient.Jar.Cookies(u) {
		req.AddCookie(v)
	}
CLIENTDO1:
	resp, err := client.Do(req)
	if dealErr(err) != nil {
		logger.WriteLog("retrying...")
		time.Sleep(1 * time.Second)
		goto CLIENTDO1
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	dealErr(err)
	return respBody

}

// DoPost Do post request returns body
func (c *Client) DoPost(url string, req *http.Request, params Params, body []byte) []byte {
	client := c.CoreClient
	u, _ := url2.Parse("http://localhost:3000/")
	for _, v := range c.CoreClient.Jar.Cookies(u) {
		req.AddCookie(v)
	}
CLIENTDO1:

	resp, err := client.Do(req)
	if dealErr(err) != nil {
		logger.WriteLog("retrying...")
		time.Sleep(1 * time.Second)
		goto CLIENTDO1
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	dealErr(err)
	return respBody

}

// dealErr Deal with some errors
func dealErr(err error) error {
	if err != nil {
		logger.WriteLog(err.Error())
		//panic(err)
		return err
	}
	return nil
}

// makeUrl format url with params
func makeUrl(urlBase string, params map[string]string) string {
	url := urlBase
	strings.TrimRight(url, "/")
	count := 0
	for k, v := range params {
		if count == 0 {
			url += fmt.Sprintf("?%s=%s", k, v)
		} else {
			url += fmt.Sprintf("&%s=%s", k, v)
		}
		count++
	}
	return url
}

func (c *Client) SaveJar(jar http.CookieJar) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		logger.WriteLog("Failed to get home path")
		return
	}
	file, err := os.OpenFile(homedir+"/xzp/cookie", os.O_WRONLY|os.O_CREATE, 0755)

	dealErr(err)
	defer file.Close()
	url, err := url2.Parse("http://localhost:3000/")
	dealErr(err)
	for _, cookie := range jar.Cookies(url) {
		fmt.Fprintf(file, "%s;", cookie.String())
	}
}

func (c *Client) LoadJar(src string) {
	file, err := os.OpenFile(src, os.O_RDONLY, 0755)
	dealErr(err)
	defer file.Close()
	cookieStr, err := ioutil.ReadAll(file)
	dealErr(err)
	vars := strings.Split(string(cookieStr), ";")
	var cookies []*http.Cookie
	day, err := time.ParseDuration("24h")
	dealErr(err)
	expireDay := time.Now().Add(day * 30)
	for i, v := range vars {
		if i == len(vars)-1 {
			break
		}
		kv := strings.Split(v, "=")
		cookies = append(cookies, &http.Cookie{
			Name:    kv[0],
			Value:   kv[1],
			Path:    "/",
			Domain:  "localhost",
			Expires: expireDay,
		})
	}
	url, err := url2.Parse("http://localhost:3000/")
	dealErr(err)
	c.CoreClient.Jar.SetCookies(url, cookies)
}

func LoadJar(src string) http.CookieJar {
	file, err := os.OpenFile(src, os.O_RDONLY, 0755)
	dealErr(err)
	defer file.Close()
	cookieStr, err := ioutil.ReadAll(file)
	dealErr(err)
	vars := strings.Split(string(cookieStr), ";")
	var cookies []*http.Cookie
	day, err := time.ParseDuration("24h")
	dealErr(err)
	expireDay := time.Now().Add(day * 30)
	for i, v := range vars {
		if i == len(vars)-1 {
			break
		}
		kv := strings.Split(v, "=")
		cookies = append(cookies, &http.Cookie{
			Name:    kv[0],
			Value:   kv[1],
			Path:    "/",
			Domain:  "localhost",
			Expires: expireDay,
		})
	}
	url, err := url2.Parse("http://localhost:3000/")
	dealErr(err)
	var cj http.CookieJar
	cj.SetCookies(url, cookies)
	return cj
}
