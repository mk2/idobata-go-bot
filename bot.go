package idobot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/r3labs/sse"
)

type OnStartHandler func(bot *Bot, msg *SeedMsg)

type OnEventHandler func(bot *Bot, msg *EventMsg)

type Bot struct {
	url       string
	botID     int
	botName   string
	client    *sse.Client
	apiToken  string
	userAgent string
	prevMsgID int
	onStart   OnStartHandler
	onEvent   OnEventHandler
}

type SeedMsg struct {
	Records struct {
		Bot struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"bot"`
	} `json:"records"`
	Version int `json:"version"`
}

type EventMsg struct {
	Data struct {
		Type    string `json:"type"`
		Message struct {
			ID            int    `json:"id"`
			Body          string `json:"body"`
			RoomID        int    `json:"room_id"`
			SenderID      int    `json:"sender_id"`
			BodyPlain     string `json:"body_plain"`
			CreatedAt     string `json:"created_at"`
			SenderName    string `json:"sender_name"`
			SenderType    string `json:"sender_type"`
			SenderIconURL string `json:"sender_icon_url"`
			Mentions      []int  `json:"mentions"`
		} `json:"message"`
	} `json:"data"`
}

// IdobataMsgFormat Idobata投稿用のメッセージフォーマット
type IdobataMsgFormat struct {
	Source string `json:"source"`
	RoomID int    `json:"room_id"`
	Format string `json:"format"`
}

// 配列に指定の数字が含まれているか確認する
func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func NewBot(url string, apiToken string, userAgent string, onStart OnStartHandler, onEvent OnEventHandler) (*Bot, error) {
	accessUrl := fmt.Sprintf("%s/api/stream?access_token=%s", url, apiToken)
	client := sse.NewClient(accessUrl)
	bot := &Bot{
		url:       url,
		client:    client,
		apiToken:  apiToken,
		onStart:   onStart,
		onEvent:   onEvent,
		userAgent: userAgent,
		botID:     -1,
	}
	for k, v := range bot.getHeaders() {
		bot.client.Headers[k] = v
	}
	return bot, nil
}

func (bot *Bot) Start() {
	bot.client.SubscribeRaw(func(evt *sse.Event) {
		if bot.botID == -1 {
			// botIDが設定されていないので、初回メッセージ
			var msg SeedMsg
			if err := json.Unmarshal(evt.Data, &msg); err != nil {
				log.Fatal(err)
			}
			bot.botID = msg.Records.Bot.ID
			bot.botName = msg.Records.Bot.Name
			fmt.Printf("botID=%d, botName=%s\n", bot.botID, bot.botName)
			bot.onStart(bot, &msg)
		} else {
			// 普通のイベントを受け取る
			var msg EventMsg
			if err := json.Unmarshal(evt.Data, &msg); err != nil {
				log.Fatal(err)
			}

			// Bot宛のメッセージは使わない
			// 無限ループするので
			if msg.Data.Message.SenderType == "Bot" {
				return
			}

			// メンションも扱わない
			if !contains(msg.Data.Message.Mentions, bot.botID) {
				return
			}

			if bot.prevMsgID != msg.Data.Message.ID {
				bot.prevMsgID = msg.Data.Message.ID
				bot.onEvent(bot, &msg)
			}
		}
	})
}

func (bot *Bot) getHeaders() map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", bot.apiToken),
		"User-Agent":    bot.userAgent,
	}
}

func (bot *Bot) PostMessage(roomID int, message string) (string, error) {
	if roomID == 0 {
		return "", nil
	}

	var (
		req  *http.Request
		res  *http.Response
		err  error
		form []byte
	)

	postMsg := &IdobataMsgFormat{
		Source: message,
		RoomID: roomID,
		Format: "html",
	}

	if form, err = json.Marshal(postMsg); err != nil {
		return "", err
	}

	client := &http.Client{}
	url := fmt.Sprintf("%s/api/messages", bot.url)
	if req, err = http.NewRequest("POST", url, strings.NewReader(string(form))); err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range bot.getHeaders() {
		req.Header.Set(k, v)
	}

	if res, err = client.Do(req); err != nil {
		return "", err
	}

	bodyBytes, _ := ioutil.ReadAll(res.Body)
	bodyString := string(bodyBytes)

	defer res.Body.Close()

	return bodyString, nil
}

func (bot *Bot) BotID() int {
	return bot.botID
}

func (bot *Bot) BotName() string {
	return bot.botName
}
