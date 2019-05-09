package main

import (
	"fmt"
	"log"
	"os"

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

func onStart(bot idobot.Bot, msg *idobot.SeedMsg) {
	fmt.Printf("[%s] Connection Established.\n", bot.BotName())
}

func onEvent(bot idobot.Bot, msg *idobot.EventMsg) {
	fireEventHandlers(bot, msg)
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
