package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mk2/idobot"
)

func fireEventHandlers(bot idobot.Bot, msg *idobot.EventMsg) {
	botName := bot.BotName()
	roomID := msg.Data.Message.RoomID
	message := string([]rune(msg.Data.Message.BodyPlain)[len(botName)+2:])
	fmt.Printf("[%s][%d] Message: %s\n", botName, roomID, message)

	switch {
	case strings.HasPrefix(message, "hi"):
		onHiEvent(bot, msg)
	case strings.HasPrefix(message, "lot"):
		onLotteryEvent(bot, msg)
	default:
		onProgramEvent(bot, msg)
	}
}

// onHiEvent `@gohper hi`で始まるイベントを処理する
func onHiEvent(bot idobot.Bot, msg *idobot.EventMsg) {
	roomID := msg.Data.Message.RoomID
	senderName := msg.Data.Message.SenderName
	bot.PostMessage(roomID, fmt.Sprintf("%s、こんにちはだGO。", senderName))
}

// onLotteryEvent `@gopher lot`で始まるイベントを処理する
func onLotteryEvent(bot idobot.Bot, msg *idobot.EventMsg) {
	botNameLen := len(bot.BotName())
	roomID := msg.Data.Message.RoomID
	rawMessage := string([]rune(msg.Data.Message.BodyPlain))

	if len(rawMessage) < (botNameLen + 6) {
		bot.PostMessage(roomID, "ちゃんと使ってGo。")
		return
	}

	message := string([]rune(msg.Data.Message.BodyPlain)[botNameLen+6:])

	rangePtn := regexp.MustCompile(`-?[\d]+\.\.-?[\d]+`)

	rand.Seed(time.Now().UnixNano())
	if rangePtn.MatchString(message) {
		tokens := strings.Split(message, "..")
		if len(tokens) < 2 {
			bot.PostMessage(roomID, fmt.Sprintf("`1..10` の形で指定してGo。"))
			return
		}
		startNum, err := strconv.Atoi(strings.TrimSpace(tokens[0]))
		endNum, err := strconv.Atoi(strings.TrimSpace(tokens[1]))

		if err != nil {
			bot.PostMessage(roomID, fmt.Sprintf("`1..10` の形で指定してGo。"))
			return
		}

		if startNum > endNum {
			startNum, endNum = endNum, startNum
		}

		nums := make([]int, endNum-startNum+1)
		for i := range nums {
			nums[i] = startNum + i
		}
		num := nums[rand.Intn(len(nums))]
		bot.PostMessage(roomID, fmt.Sprintf("%d", num))
	} else {
		tokens := strings.Split(message, " ")
		token := tokens[rand.Intn(len(tokens))]
		bot.PostMessage(roomID, fmt.Sprintf("%s", token))
	}
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
