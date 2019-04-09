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

func main() {
	idobataAPIToken := os.Getenv("IDOBATA_API_TOKEN")
	if len(idobataAPIToken) == 0 {
		log.Fatal("IDOBATA_API_TOKEN was not set.")
	}

	var onStart idobot.OnStartHandler = func(bot idobot.Bot, msg *idobot.SeedMsg) {
		fmt.Printf("[%s] Connection Established.\n", bot.BotName())
	}

	var onEvent idobot.OnEventHandler = func(bot idobot.Bot, msg *idobot.EventMsg) {
		botName := bot.BotName()
		roomID := msg.Data.Message.RoomID
		messageBody := msg.Data.Message.BodyPlain
		fmt.Printf("[%s][%d] Message: %s\n", botName, roomID, messageBody)
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

	var onError idobot.OnErrorHandler = func(bot idobot.Bot, err error) {
		for roomID := range bot.RoomIDs() {
			errMsg := fmt.Sprintf("エラー: %+v \n%sは終了しますGO。再起動してGO。", err, bot.BotName())
			bot.PostMessage(roomID, errMsg)
		}
	}

	bot, err := idobot.NewBot(idobataURL, idobataAPIToken, userAgent, onStart, onEvent, onError)
	if err != nil {
		log.Fatal("Bot cannot be initialized.")
	}
	fmt.Println("Bot was initialized.")
	bot.Start()
}
