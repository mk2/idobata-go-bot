package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/mk2/idobot"
)

const (
	idobataURL    = "https://idobata.io"
	userAgent     = "gopher-bot"
	playGroundURL = "https://play.golang.org"
)

// FormatResp Playgroundからのレスポンス
type FormatResp struct {
	Body  string
	Error string
}

// CompileResp Playgroundからのレスポンス
type CompileResp struct {
	Errors string
	Events []struct {
		Message string
		Kind    string
		Delay   int
	}
	Status      int
	IsTest      bool
	TestsFailed int
}

func request(baseURL string, path string, userAgent string, payload map[string]string, output interface{}) (string, error) {
	values := url.Values{}
	for k, v := range payload {
		values.Set(k, v)
	}
	reqURL := fmt.Sprintf("%s/%s", baseURL, path)
	req, err := http.NewRequest("POST", reqURL, strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("User-Agent", userAgent)
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	bodyString := string(bodyBytes)
	err = json.Unmarshal(bodyBytes, output)
	return bodyString, err
}

func formatProgram(program string) (*FormatResp, error) {
	var output FormatResp
	_, err := request(playGroundURL, "fmt", userAgent, map[string]string{
		"imports": "true",
		"body":    program,
	}, &output)
	if err != nil {
		return nil, err
	}
	if len(output.Error) > 0 {
		return nil, errors.New(output.Error)
	}
	return &output, nil
}

func runProgram(program string) (*CompileResp, error) {
	var output CompileResp
	_, err := request(playGroundURL, "compile", userAgent, map[string]string{
		"version": "2",
		"body":    program,
	}, &output)
	if err != nil {
		return nil, err
	}

	if len(output.Events) < 1 {
		return nil, errors.New(output.Errors)
	}

	return &output, nil
}

func onStart(bot idobot.Bot, msg *idobot.SeedMsg) {
	fmt.Printf("[%s] Connection Established.\n", bot.BotName())
}

// onHiEvent `@gohper hi`で始まるイベントを処理する
func onHiEvent(bot idobot.Bot, msg *idobot.EventMsg) {
	roomID := msg.Data.Message.RoomID
	senderName := msg.Data.Message.SenderName
	bot.PostMessage(roomID, fmt.Sprintf("%s、こんにちはだGO。", senderName))
}

// onProgramEvent デフォルトのイベントハンドラ
func onProgramEvent(bot idobot.Bot, msg *idobot.EventMsg) {
	botName := bot.BotName()
	roomID := msg.Data.Message.RoomID
	messageBody := msg.Data.Message.BodyPlain

	program := string([]rune(messageBody)[len(botName)+2:])
	compileOutput, err := runProgram(program)
	fmt.Printf("[%s] program=%s\n", bot.BotName(), program)
	if err != nil {
		bot.PostMessage(roomID, fmt.Sprintf("エラー[compile]: %s", err))
		return
	}
	formatOutput, err := formatProgram(program)
	if err != nil {
		bot.PostMessage(roomID, fmt.Sprintf("エラー[format]: %s", err))
		return
	}
	message := fmt.Sprintf("<p><pre>%s</pre></p><details><summary>元のプログラム</summary><pre>%s</pre></details>", compileOutput.Events[0].Message, formatOutput.Body)
	fmt.Println(message)
	bot.PostMessage(roomID, message)
}

func onEvent(bot idobot.Bot, msg *idobot.EventMsg) {
	botName := bot.BotName()
	roomID := msg.Data.Message.RoomID
	message := string([]rune(msg.Data.Message.BodyPlain)[len(botName)+2:])
	fmt.Printf("[%s][%d] Message: %s\n", botName, roomID, message)

	switch {
	case strings.HasPrefix(message, "hi"):
		onHiEvent(bot, msg)
	default:
		onProgramEvent(bot, msg)
	}
}

func onError(bot idobot.Bot, err error) {
	for roomID := range bot.RoomIDs() {
		errMsg := fmt.Sprintf("エラー: %+v \n%sは終了しますGO。再起動してGO。", err, bot.BotName())
		bot.PostMessage(roomID, errMsg)
	}
}

func main() {
	idobataAPIToken := os.Getenv("IDOBATA_API_TOKEN")
	storeFilePath := os.Getenv("STORE_FILE_PATH")
	if len(idobataAPIToken) == 0 {
		log.Fatal("IDOBATA_API_TOKEN was not set.")
	}

	if len(storeFilePath) == 0 {
		log.Fatal("STORE_FILE_PATH was not set.")
	}

	bot, err := idobot.NewBot(&idobot.NewBotOpts{
		URL:           idobataURL,
		APIToken:      idobataAPIToken,
		UserAgent:     userAgent,
		StoreFilePath: storeFilePath,
		OnStart:       onStart,
		OnEvent:       onEvent,
		OnError:       onError,
	})
	if err != nil {
		log.Fatal("Bot cannot be initialized.")
	}
	fmt.Println("Bot was initialized.")
	bot.Run()
}
