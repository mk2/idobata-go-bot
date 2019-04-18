package idobot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/r3labs/sse"
	bolt "go.etcd.io/bbolt"
)

// Bot idobotを使うプログラムから、idobataへアクセスする方法
type Bot interface {
	Start() error
	Stop() error
	PostMessage(roomID int, message string) (string, error)
	RoomIDs() []int
	BotID() int
	BotName() string
	DB() *bolt.DB
	PutDB(key, value string) error
	GetDB(key string) (string, error)
}

// OnStartHandler idobot開始時に呼ばれるコールバック
type OnStartHandler func(bot Bot, msg *SeedMsg)

// OnEventHandler idobotがメッセージを受信した際に呼ばれるコールバック
type OnEventHandler func(bot Bot, msg *EventMsg)

// OnErrorHandler idobotが何かしらのエラーに遭遇した際に呼ばれるコールバック
type OnErrorHandler func(bot Bot, err error)

// SeedMsg 開始時に送られてくるメッセージ
// botの名前やidが含まれる
type SeedMsg struct {
	Records struct {
		Bot struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"bot"`
	} `json:"records"`
	Version int `json:"version"`
}

// EventMsg 通常時に受信するメッセージ
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

// roomIDSet idobotが動いている最中に、メッセージを受信した部屋ID一覧
type roomIDSet struct {
	set map[int]bool
}

func (set *roomIDSet) add(i int) bool {
	_, found := set.set[i]
	set.set[i] = true
	return !found
}

// botImpl Botインターフェースを実装している実体
type botImpl struct {
	Bot
	url        string
	botID      int
	botName    string
	client     *sse.Client
	apiToken   string
	userAgent  string
	prevMsgID  int
	onStart    OnStartHandler
	onEvent    OnEventHandler
	onError    OnErrorHandler
	roomIDs    *roomIDSet
	db         *bolt.DB
	bucketName string
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

// NewBotOpts 新しくBotを作るときのオプション
type NewBotOpts struct {
	URL        string
	APIToken   string
	UserAgent  string
	BucketName string
	DBName     string
	OnStart    OnStartHandler
	OnEvent    OnEventHandler
	OnError    OnErrorHandler
}

// NewBot 新しくidobotを作る
func NewBot(opts *NewBotOpts) (Bot, error) {
	// バケット名を決める
	bucketName := "User"
	if opts.BucketName != "" {
		bucketName = opts.BucketName
	}
	// boltDBファイルを作成
	db, err := bolt.Open(opts.DBName, 0600, nil)
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(bucketName))
		return err
	})
	if err != nil {
		return nil, err
	}
	accessURL := fmt.Sprintf("%s/api/stream?access_token=%s", opts.URL, opts.APIToken)
	client := sse.NewClient(accessURL)
	bot := &botImpl{
		url:        opts.URL,
		client:     client,
		apiToken:   opts.APIToken,
		onStart:    opts.OnStart,
		onEvent:    opts.OnEvent,
		onError:    opts.OnError,
		userAgent:  opts.UserAgent,
		botID:      -1,
		roomIDs:    &roomIDSet{set: make(map[int]bool)},
		db:         db,
		bucketName: bucketName,
	}
	for k, v := range bot.getHeaders() {
		bot.client.Headers[k] = v
	}
	return bot, nil
}

func (bot *botImpl) Start() error {
	err := bot.client.SubscribeRaw(func(evt *sse.Event) {
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

			// RoomIDは保存しておく
			bot.roomIDs.add(msg.Data.Message.RoomID)

			if bot.prevMsgID != msg.Data.Message.ID {
				bot.prevMsgID = msg.Data.Message.ID
				bot.onEvent(bot, &msg)
			}
		}
	})

	bot.onError(bot, err)

	return err
}

func (bot *botImpl) Stop() error {
	err := bot.db.Close()
	return err
}

func (bot *botImpl) getHeaders() map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", bot.apiToken),
		"User-Agent":    bot.userAgent,
	}
}

func (bot *botImpl) PostMessage(roomID int, message string) (string, error) {
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
