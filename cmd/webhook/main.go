package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/line/line-bot-sdk-go/linebot"
)

var bot *linebot.Client

var applog *logger

var chaSec string

func main() {
	s := flag.String("s", "ChannelSecret", "ChannelSecret")
	t := flag.String("t", "AccessToken", "AccessToken")
	flag.Parse()

	logfile, err := SetupLog("/var/log")
	if err != nil {
		os.Exit(1)
	}
	defer logfile.Close()

	applog = &logger{isDebugEnable: true}
	applog.debug("&&&&&&&&&&&&&&&&&&&&&&&&&")
	applog.debug("START")
	applog.debug(*s)
	applog.debug(*t)
	applog.debug("&&&&&&&&&&&&&&&&&&&&&&&&&")

	chaSec = *s
	bot, err = linebot.New(chaSec, *t)
	if err != nil {
		applog.errorf("[main]", "Err: %+v", err)
		return
	}

	http.HandleFunc("/srr/webhook", srrHandler)
	if err := http.ListenAndServe(":20051", nil); err != nil {
		applog.errorf("[main]", "Err: %+v", err)
	}
}

func srrHandler(w http.ResponseWriter, r *http.Request) {
	events, err := ParseRequest(chaSec, r)
	if err != nil {
		applog.errorf("[srrHandler]", "Err: %+v", err)
	}

	applog.debug("[srrHandler]", "Event for start")

	for _, event := range events {
		applog.debugf("[srrHandler]", "Event Type: %+v", event.Type)
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				applog.debugf("[srrHandler]", "Event Message Type: %+v", message)
				var newMsg *linebot.TextMessage
				if "あぶない" == message.Text {
					newMsg = linebot.NewTextMessage("ばしょをちずでおしえて！")
				} else {
					newMsg = linebot.NewTextMessage(message.Text + "!?")
				}
				applog.debugf("[srrHandler]", "newMsg: %s", newMsg)

				repMsg := bot.ReplyMessage(event.ReplyToken, newMsg)
				applog.debugf("[srrHandler]", "replyMessage: %+v", repMsg)
				if _, err = bot.ReplyMessage(event.ReplyToken, newMsg).Do(); err != nil {
					applog.errorf("[srrHandler]", "Err: %+v", err)
				}
			case *linebot.LocationMessage:
				applog.debugf("[srrHandler]", "Event Message Type: %+v", message)
				lat := message.Latitude
				lon := message.Longitude
				addr := message.Address
				retMsg := fmt.Sprintf("じゅうしょは、%s \n緯度：%f\n経度：%f\nだね。ありがとう。みんなにもおしえてあげるね。", addr, lat, lon)
				applog.debugf("[srrHandler]", "retMsg: %s", retMsg)
				newMsg := linebot.NewTextMessage(retMsg)
				repMsg := bot.ReplyMessage(event.ReplyToken, newMsg)
				applog.errorf("[srrHandler]", "replyMessage: %+v", repMsg)
				if _, err = bot.ReplyMessage(event.ReplyToken, newMsg).Do(); err != nil {
					applog.errorf("[srrHandler]", "Err: %+v", err)
				}
			}
		}
	}

}

// ParseRequest func
func ParseRequest(channelSecret string, r *http.Request) ([]*linebot.Event, error) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	applog.debugf("[ParseRequest]", "channelSecret: %s", channelSecret)

	lineSign := r.Header.Get("X-Line-Signature")
	applog.debugf("[ParseRequest]", "X-Line-Signature: %s", lineSign)
	if !validateSignature(channelSecret, lineSign, body) {
		return nil, linebot.ErrInvalidSignature
	}
	applog.debug("[ParseRequest]", "valiate ok")

	request := &struct {
		Events []*linebot.Event `json:"events"`
	}{}
	if err = json.Unmarshal(body, request); err != nil {
		return nil, err
	}
	applog.debugf("[ParseRequest]", "==##==##==##==##==##==##==##==##==")
	applog.debugf("[ParseRequest]", "request: %+v", request)
	applog.debugf("[ParseRequest]", "==##==##==##==##==##==##==##==##==")

	return request.Events, nil
}

func validateSignature(channelSecret, signature string, body []byte) bool {
	decoded, err := base64.StdEncoding.DecodeString(signature)
	applog.debugf("[validateSignature]", "decoded: %s", decoded)
	if err != nil {
		return false
	}
	hash := hmac.New(sha256.New, []byte(channelSecret))
	hash.Write(body)
	applog.debugf("[validateSignature]", "hash.Sum(nil): %s", hash.Sum(nil))
	return hmac.Equal(decoded, hash.Sum(nil))
}

func srrHandler2(w http.ResponseWriter, r *http.Request) {
	ba, err := ioutil.ReadAll(r.Body)
	if err != nil {
		applog.errorf("[srrHandler]", "Err: %+v", err)
	} else {
		applog.debugf("[srrHandler]", "X-Line-Signature: ", r.Header.Get("X-Line-Signature"))
		for k, v := range r.Header {
			applog.debugf("[srrHandler]", "RequestHeader: key[%+v], value[%+v]\n", k, v)
		}
		applog.debugf("[srrHandler]", "RequestBody: %+v", string(ba))
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		applog.errorf("[srrHandler]", "Err: %+v", err)
	}
	applog.debugf("[srrHandler]", "$$$ RequestBody $$$ : %+v", body)

	events, err := bot.ParseRequest(r)
	// defer r.Body.Close()
	// body, err := ioutil.ReadAll(r.Body)
	// if err != nil {
	// 	applog.errorf("[srrHandler]", "Err: %+v", err)
	// }
	// applog.errorf("[srrHandler]", "channelSecret: x%+vs", r.Header.Get("X-Line-Signature"))
	// if !validateSignature(channelSecret, r.Header.Get("X-Line-Signature"), body) {
	// 	return nil, ErrInvalidSignature
	// }

	// request := &struct {
	// 	Events []*linebot.Event `json:"events"`
	// }{}
	// if err = json.Unmarshal(body, request); err != nil {
	// 	applog.errorf("[srrHandler]", "Err: %+v", err)
	// }
	// events, err := request.Events, nil

	if err != nil {
		applog.errorf("[srrHandler]", "Err: %+v", err)
		// if err == linebot.ErrInvalidSignature {
		// 	w.WriteHeader(400)
		// } else {
		// 	w.WriteHeader(500)
		// }
		// return
	}

	applog.debug("[srrHandler]", "Event for start")

	for _, event := range events {
		applog.debugf("[srrHandler]", "Event Type: %+v", event.Type)
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				applog.debugf("[srrHandler]", "Event Message Type: %+v", message)
				var newMsg *linebot.TextMessage
				if "あぶない" == message.Text {
					newMsg = linebot.NewTextMessage("ばしょをちずでおしえて！")
				} else {
					newMsg = linebot.NewTextMessage(message.Text + "!?")
				}
				applog.debugf("[srrHandler]", "newMsg: %s", newMsg)

				repMsg := bot.ReplyMessage(event.ReplyToken, newMsg)
				applog.errorf("[srrHandler]", "replyMessage: %+v", repMsg)
				if _, err = bot.ReplyMessage(event.ReplyToken, newMsg).Do(); err != nil {
					applog.errorf("[srrHandler]", "Err: %+v", err)
				}
			case *linebot.LocationMessage:
				applog.debugf("[srrHandler]", "Event Message Type: %+v", message)
				lat := message.Latitude
				lon := message.Longitude
				addr := message.Address
				retMsg := fmt.Sprintf("じゅうしょは、%s \n緯度：%f\n経度：%f\nだね。ありがとう。みんなにもおしえてあげるね。", addr, lat, lon)
				applog.debugf("[srrHandler]", "retMsg: %s", retMsg)
				newMsg := linebot.NewTextMessage(retMsg)
				repMsg := bot.ReplyMessage(event.ReplyToken, newMsg)
				applog.errorf("[srrHandler]", "replyMessage: %+v", repMsg)
				if _, err = bot.ReplyMessage(event.ReplyToken, newMsg).Do(); err != nil {
					applog.errorf("[srrHandler]", "Err: %+v", err)
				}
			}
		}
	}

}

// SetupLog ...
func SetupLog(outputDir string) (*os.File, error) {
	logfile, err := os.OpenFile(filepath.Join(outputDir, "srr-line-webhook.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Printf("[%s]のログファイル「srr-line-webhook.log」オープンに失敗しました。 [ERROR]%s\n", outputDir, err)
		return nil, err
	}

	// [MEMO]内容に応じて出力するファイルを切り替える場合はどうするんだ・・・？
	log.SetOutput(io.MultiWriter(logfile, os.Stdout))
	log.SetFlags(log.Ldate | log.Ltime)

	return logfile, nil
}

// [MEMO]セットアップ時に「log.Lshortfile」セットしたかったけど、このファイルでログ出力を担うようにすると全ログの「log.Lshortfile」結果が「logger.go」になるので諦め
// TODO ロギングフレームワークを採用する！　logrusあたりがメジャーらしい。最近では（ベータ版だけど）zapがいいらしい。
type logger struct {
	isDebugEnable bool
}

func (l *logger) debug(fname string, v ...interface{}) {
	if l.isDebugEnable {
		log.Println("[DEBUG]["+fname+"]", v)
	}
}

func (l *logger) debugf(fname string, formatStr string, v ...interface{}) {
	if l.isDebugEnable {
		log.Printf("[DEBUG]["+fname+"] "+formatStr, v)
	}
}

func (l *logger) info(fname string, v ...interface{}) {
	log.Println("[INFO]["+fname+"]", v)
}

func (l *logger) infof(fname string, formatStr string, v ...interface{}) {
	log.Printf("[INFO]["+fname+"] "+formatStr, v)
}

func (l *logger) error(fname string, v ...interface{}) {
	log.Println("[ERROR]["+fname+"]", v)
}

func (l *logger) errorf(fname string, formatStr string, v ...interface{}) {
	log.Printf("[ERROR]["+fname+"] "+formatStr, v)
}

// Message EventのTextとLocationを扱う。
// Textはキーワードリターン用
// Locationはストレージ保存用

// のち、PUB/SUBにする。
