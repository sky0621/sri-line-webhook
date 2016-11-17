package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/line/line-bot-sdk-go/linebot"
)

var bot *linebot.Client

func main() {
	s := flag.String("s", "ChannelSecret", "ChannelSecret")
	t := flag.String("t", "AccessToken", "AccessToken")
	flag.Parse()

	var err error
	bot, err = linebot.New(*s, *t)
	if err != nil {
		log.Println(err)
		return
	}

	http.HandleFunc("/srr/webhook", srrHandler)
	if err := http.ListenAndServe(":20051", nil); err != nil {
		log.Println(err)
	}
}

func srrHandler(w http.ResponseWriter, r *http.Request) {
	events, err := bot.ParseRequest(r)
	if err != nil {
		log.Println(err)
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				log.Println(message)
				newMsg := linebot.NewTextMessage(message.Text + "!?")
				if _, err = bot.ReplyMessage(event.ReplyToken, newMsg).Do(); err != nil {
					log.Println(err)
				}
			}
		}
	}

}

// Message EventのTextとLocationを扱う。
// Textはキーワードリターン用
// Locationはストレージ保存用

// のち、PUB/SUBにする。
