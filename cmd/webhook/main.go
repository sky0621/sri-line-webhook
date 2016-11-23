package main

import (
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
	applog.debug("START")

	bot, err = linebot.New(*s, *t)
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
	ba, err := ioutil.ReadAll(r.Body)
	if err != nil {
		applog.errorf("[srrHandler]", "Err: %+v", err)
	} else {
		for k, v := range r.Header {
			applog.debugf("[srrHandler]", "RequestHeader: key[%+v], value[%+v]\n", k, v)
		}
		applog.debugf("[srrHandler]", "RequestBody: %+v", string(ba))
	}
	events, err := bot.ParseRequest(r)
	if err != nil {
		applog.errorf("[srrHandler]", "Err: %+v", err)
		// if err == linebot.ErrInvalidSignature {
		// 	w.WriteHeader(400)
		// } else {
		// 	w.WriteHeader(500)
		// }
		// return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				log.Println(message)
				var newMsg *linebot.TextMessage
				if "あぶない" == message.Text {
					newMsg = linebot.NewTextMessage("ばしょをちずでおしえて！")
				} else {
					newMsg = linebot.NewTextMessage(message.Text + "!?")
				}
				applog.debugf("[srrHandler]", "newMsg: %s", newMsg)
				if _, err = bot.ReplyMessage(event.ReplyToken, newMsg).Do(); err != nil {
					applog.errorf("[srrHandler]", "Err: %+v", err)
				}
			case *linebot.LocationMessage:
				log.Println(message)
				lat := message.Latitude
				lon := message.Longitude
				addr := message.Address
				retMsg := fmt.Sprintf("じゅうしょは、%s \n緯度：%f\n経度：%f\nだね。ありがとう。みんなにもおしえてあげるね。", addr, lat, lon)
				applog.debugf("[srrHandler]", "retMsg: %s", retMsg)
				newMsg := linebot.NewTextMessage(retMsg)
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
