package main

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/eddieivan01/nic"
	"github.com/google/go-github/github"
	jsoniter "github.com/json-iterator/go"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
	"log"
	math_rand "math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var session = nic.Session{}

func main() {
	conf := os.Getenv("CONFIG")
	config := ReadYamlConfig(conf)
	run(config)
}
func run(config *Config) {
	cookie := login(config.UserPwd)
	fns := GetFiles(config.RootFileId, config.RootFileId)
	json, _ := jsoniter.MarshalToString(fns)
	if len(fns) > 0 {
		log.Println(">> 数据获取成功：" + strconv.Itoa(len(json)))
	} else {
		log.Println(">> 数据获取失败：" + strconv.Itoa(len(json)))
	}
	//ioutil.WriteFile("/home/single/Desktop/data.json", []byte(json), 0644)
	pushToGithub("data/data.json", json, config.GhToken)
	//签到、抽奖
	DayTask(cookie)
}
func DayTask(cookie string) {
	rand := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
	surl := "https://api.cloud.189.cn/mkt/userSign.action?rand=" + rand + "&clientType=TELEANDROID&version=8.6.3&model=SM-G930K"
	url := "https://m.cloud.189.cn/v2/drawPrizeMarketDetails.action?taskId=TASK_SIGNIN&activityId=ACT_SIGNIN"
	url2 := "https://m.cloud.189.cn/v2/drawPrizeMarketDetails.action?taskId=TASK_SIGNIN_PHOTOS&activityId=ACT_SIGNIN"
	resp, err := nic.Get(surl, nic.H{
		Cookies: nic.KV{
			"COOKIE_LOGIN_USER": cookie,
		},
		Headers: nic.KV{
			"User-Agent":      "Mozilla/5.0 (Linux; Android 5.1.1; SM-G930K Build/NRD90M; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/74.0.3729.136 Mobile Safari/537.36 Ecloud/8.6.3 Android/22 clientId/355325117317828 clientModel/SM-G930K imsi/460071114317824 clientChannelId/qq proVersion/1.0.6",
			"Referer":         "https://m.cloud.189.cn/zhuanti/2016/sign/index.jsp?albumBackupOpened=1",
			"Host":            "m.cloud.189.cn",
			"Accept-Encoding": "gzip, deflate",
		},
	})
	netdiskBonus := jsoniter.Get([]byte(resp.Text), "netdiskBonus").ToString()
	if err != nil {
		log.Fatal(err.Error())
	}
	if jsoniter.Get([]byte(resp.Text), "isSign").ToString() == "false" {
		log.Println(">> 未签到，签到获得" + netdiskBonus + "M空间")
	} else {
		log.Println(">> 已经签到过了，签到获得" + netdiskBonus + "M空间")
	}
	headers := nic.KV{
		"User-Agent":      "Mozilla/5.0 (Linux; Android 5.1.1; SM-G930K Build/NRD90M; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/74.0.3729.136 Mobile Safari/537.36 Ecloud/8.6.3 Android/22 clientId/355325117317828 clientModel/SM-G930K imsi/460071114317824 clientChannelId/qq proVersion/1.0.6",
		"Referer":         "https://m.cloud.189.cn/zhuanti/2016/sign/index.jsp?albumBackupOpened=1",
		"Host":            "m.cloud.189.cn",
		"Accept-Encoding": "gzip, deflate",
	}
	resp2, _ := nic.Get(url, nic.H{
		Cookies: nic.KV{
			"COOKIE_LOGIN_USER": cookie,
		},
		Headers: headers,
	})
	if strings.Contains(resp2.Text, "User_Not_Chance") {
		log.Println(">> 已经抽奖过了，今天已经没有机会")
	} else {
		description := jsoniter.Get([]byte(resp2.Text), "description").ToString()
		log.Println(">> 抽奖获得" + description)
	}
	resp3, _ := nic.Get(url2, nic.H{
		Cookies: nic.KV{
			"COOKIE_LOGIN_USER": cookie,
		},
		Headers: headers,
	})
	if strings.Contains(resp3.Text, "User_Not_Chance") {
		log.Println(">> 已经抽奖过了，今天已经没有机会")
	} else {

		description := jsoniter.Get([]byte(resp3.Text), "description").ToString()
		log.Println(">> 抽奖获得" + description)
	}
}
func random() string {
	return fmt.Sprintf("0.%17v", math_rand.New(math_rand.NewSource(time.Now().UnixNano())).Int63n(100000000000000000))
}
func GetFiles(rootId, fileId string) []FileNode {
	fns := make([]FileNode, 0)
	pageNum := 1
	for {
		url := fmt.Sprintf("https://cloud.189.cn/v2/listFiles.action?fileId=%s&mediaType=&keyword=&inGroupSpace=false&orderBy=3&order=DESC&pageNum=%d&pageSize=100&noCache=%s", fileId, pageNum, random())
		resp, err := session.Get(url, nil)
		if err != nil {
			log.Fatal(err.Error())
		}
		byteFiles := []byte(resp.Text)
		totalCount := jsoniter.Get(byteFiles, "recordCount").ToInt()
		d := jsoniter.Get(byteFiles, "data")
		paths := jsoniter.Get(byteFiles, "path")
		ps := []Paths{}
		err = jsoniter.Unmarshal([]byte(paths.ToString()), &ps)
		p := ""
		flag := false
		if err == nil {
			for _, item := range ps {
				if flag == true && item.FileId != rootId {
					if strings.HasSuffix(p, "/") != true {
						p += "/" + item.FileName
					} else {
						p += item.FileName
					}
				}
				if item.FileId == rootId {
					flag = true
				}
				if flag == true && item.FileId == rootId {
					p += "/"
				}
			}
		}
		if d != nil {
			m := []FileNode{}
			err = jsoniter.Unmarshal([]byte(d.ToString()), &m)
			if err == nil {
				fns = fns[:0]
				for _, item := range m {
					item.Path = p
					if item.IsFolder == true {
						item.Children = GetFiles(rootId, item.FileId)
					}
					fns = append(fns, item)
				}
			}
		}
		if pageNum*100 < totalCount {
			pageNum++
		} else {
			break
		}
	}
	return fns
}

type Paths struct {
	FileId    string `json:"fileId"`
	FileName  string `json:"fileName"`
	IsCoShare int    `json:"isCoShare"`
}
type FileNode struct {
	FileId       string     `json:"fileId"`
	FileIdDigest string     `json:"fileIdDigest"`
	FileName     string     `json:"fileName"`
	FileSize     int        `json:"fileSize"`
	FileType     string     `json:"fileType"`
	IsFolder     bool       `json:"isFolder"`
	IsStarred    bool       `json:"isStarred"`
	LastOpTime   string     `json:"lastOpTime"`
	ParentId     string     `json:"parentId"`
	Path         string     `json:"path"`
	DownloadUrl  string     `json:"downloadUrl"`
	MediaType    int        `json:"mediaType"`
	Icon         Icon       `json:"icon"`
	CreateTime   string     `json:"create_time"`
	Children     []FileNode `json:"children"`
}
type Icon struct {
	LargeUrl string `json:"largeUrl"`
	SmallUrl string `json:"smallUrl"`
}

//天翼云网盘登录
func login(userPwd string) string {
	userPwdArr := strings.Split(userPwd, " ")
	user := userPwdArr[0]
	password := userPwdArr[1]
	url := "https://cloud.189.cn/udb/udb_login.jsp?pageId=1&redirectURL=/main.action"
	res, _ := session.Get(url, nil)
	b := res.Text
	lt := regexp.MustCompile(`lt = "(.+?)"`).FindStringSubmatch(b)[1]
	captchaToken := regexp.MustCompile(`captchaToken' value='(.+?)'`).FindStringSubmatch(b)[1]
	returnUrl := regexp.MustCompile(`returnUrl = '(.+?)'`).FindStringSubmatch(b)[1]
	paramId := regexp.MustCompile(`paramId = "(.+?)"`).FindStringSubmatch(b)[1]
	//reqId := regexp.MustCompile(`reqId = "(.+?)"`).FindStringSubmatch(b)[1]
	jRsakey := regexp.MustCompile(`j_rsaKey" value="(\S+)"`).FindStringSubmatch(b)[1]
	userRsa := RsaEncode([]byte(user), jRsakey)
	passwordRsa := RsaEncode([]byte(password), jRsakey)
	url = "https://open.e.189.cn/api/logbox/oauth2/loginSubmit.do"
	loginResp, _ := session.Post(url, nic.H{
		Data: nic.KV{
			"appKey":       "cloud",
			"accountType":  "01",
			"userName":     "{RSA}" + userRsa,
			"password":     "{RSA}" + passwordRsa,
			"validateCode": "",
			"captchaToken": captchaToken,
			"returnUrl":    returnUrl,
			"mailSuffix":   "@189.cn",
			"paramId":      paramId,
			"clientType":   "10010",
			"dynamicCheck": "FALSE",
			"cb_SaveName":  "1",
			"isOauth2":     "false",
		},
		Headers: nic.KV{
			"lt":         lt,
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:74.0) Gecko/20100101 Firefox/76.0",
			"Referer":    "https://open.e.189.cn/",
		},
	})
	restCode := jsoniter.Get([]byte(loginResp.Text), "result").ToInt()
	//0登录成功，-2，需要获取验证码，-5 app info获取失败
	if restCode == 0 {
		log.Println(">> 登录成功。")
		toUrl := jsoniter.Get([]byte(loginResp.Text), "toUrl").ToString()
		res, _ := session.Get(toUrl, nil)
		return res.Cookies()[0].Value
	}
	return ""
}

func pushToGithub(path, data, token string) error {
	r := "pan-189"
	if data == "" {
		return errors.New("params error")
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	c := "同步数据：" + path
	sha := ""
	content := &github.RepositoryContentFileOptions{
		Message: &c,
		SHA:     &sha,
		Branch:  github.String("master"),
	}
	op := &github.RepositoryContentGetOptions{}
	user, _, _ := client.Users.Get(ctx, "")
	repo, _, _, er := client.Repositories.GetContents(ctx, user.GetLogin(), r, path, op)
	if er != nil || repo == nil {
		log.Println("get github repository error, create")
		content.Content = []byte(data)
		_, _, err := client.Repositories.CreateFile(ctx, user.GetLogin(), r, path, content)
		if err != nil {
			log.Println(err)
			return err
		}
	} else {
		content.SHA = repo.SHA
		content.Content = []byte(data)
		_, _, err := client.Repositories.UpdateFile(ctx, user.GetLogin(), r, path, content)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

// 加密
func RsaEncode(origData []byte, j_rsakey string) string {
	publicKey := []byte("-----BEGIN PUBLIC KEY-----\n" + j_rsakey + "\n-----END PUBLIC KEY-----")
	block, _ := pem.Decode(publicKey)
	pubInterface, _ := x509.ParsePKIXPublicKey(block.Bytes)
	pub := pubInterface.(*rsa.PublicKey)
	b, err := rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
	if err != nil {
		fmt.Println("err: " + err.Error())
	}
	return b64tohex(base64.StdEncoding.EncodeToString(b))
}

var b64map = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

var BI_RM = "0123456789abcdefghijklmnopqrstuvwxyz"

func int2char(a int) string {
	return strings.Split(BI_RM, "")[a]
}

func b64tohex(a string) string {
	d := ""
	e := 0
	c := 0
	for i := 0; i < len(a); i++ {
		m := strings.Split(a, "")[i]
		if m != "=" {
			v := strings.Index(b64map, m)
			if 0 == e {
				e = 1
				d += int2char(v >> 2)
				c = 3 & v
			} else if 1 == e {
				e = 2
				d += int2char(c<<2 | v>>4)
				c = 15 & v
			} else if 2 == e {
				e = 3
				d += int2char(c)
				d += int2char(v >> 2)
				c = 3 & v
			} else {
				e = 0
				d += int2char(c<<2 | v>>4)
				d += int2char(15 & v)
			}
		}
	}
	if e == 1 {
		d += int2char(c << 2)
	}
	return d
}

//填充字符串（末尾）
func PaddingText1(str []byte, blockSize int) []byte {
	//需要填充的数据长度
	paddingCount := blockSize - len(str)%blockSize
	//填充数据为：paddingCount ,填充的值为：paddingCount
	paddingStr := bytes.Repeat([]byte{byte(paddingCount)}, paddingCount)
	newPaddingStr := append(str, paddingStr...)
	//fmt.Println(newPaddingStr)
	return newPaddingStr
}

//去掉字符（末尾）
func UnPaddingText1(str []byte) []byte {
	n := len(str)
	count := int(str[n-1])
	newPaddingText := str[:n-count]
	return newPaddingText
}

//---------------DES加密  解密--------------------
func EncyptogAES(src, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(nil)
		return nil
	}
	src = PaddingText1(src, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key)
	blockMode.CryptBlocks(src, src)
	return src

}
func DecrptogAES(src, key []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(nil)
		return nil
	}
	blockMode := cipher.NewCBCDecrypter(block, key)
	blockMode.CryptBlocks(src, src)
	src = UnPaddingText1(src)
	return src
}

type Config struct {
	GhToken    string `yaml:"GH_TOKEN"`
	UserPwd    string `yaml:"USER_PWD"`
	RootFileId string `yaml:"ROOT_FILE_ID"`
}

func ReadYamlConfig(confText string) *Config {
	conf := &Config{}
	if confText == "" {
		return nil
	} else {
		r := strings.NewReader(confText)
		yaml.NewDecoder(r).Decode(conf)
	}
	return conf
}
