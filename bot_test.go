package idobot_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mk2/idobot"
)

func TestIdobot_新しくbotを生成できるか(t *testing.T) {
	url := "url"
	apiToken := "token"
	userAgent := "userAgent"
	onStart := func(_ idobot.Bot, _ *idobot.SeedMsg) {}
	onEvent := func(_ idobot.Bot, _ *idobot.EventMsg) {}
	onError := func(_ idobot.Bot, _ error) {}
	bot, err := idobot.NewBot(url, apiToken, userAgent, "./test.db", onStart, onEvent, onError)
	defer bot.DB().Close()
	defer os.Remove("./test.db")

	if bot == nil || err != nil {
		t.Errorf("idobot cannot instantiated with NewBot\n")
	}
}

func TestIdobot_PostMessage実行がうまくいくかどうか(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(200)
		res.Write([]byte("body"))
	}))
	url := testServer.URL
	apiToken := "token"
	userAgent := "userAgent"
	onStart := func(_ idobot.Bot, _ *idobot.SeedMsg) {}
	onEvent := func(_ idobot.Bot, _ *idobot.EventMsg) {}
	onError := func(_ idobot.Bot, _ error) {}
	bot, err := idobot.NewBot(url, apiToken, userAgent, "./test.db", onStart, onEvent, onError)
	defer bot.Stop()
	defer os.Remove("./test.db")

	if err != nil {
		t.Errorf("NewBot failed to generate bot.")
	}

	body, err := bot.PostMessage(100, "test")

	if body != "body" {
		t.Errorf("PostMessage respone: expected: \"body\"\n received: \"%s\"", body)
	}

	if err != nil {
		t.Errorf("PostMessage returns error")
	}
}

func TestIdobot_DBの書き込み読み込み(t *testing.T) {
	url := "url"
	apiToken := "token"
	userAgent := "userAgent"
	onStart := func(_ idobot.Bot, _ *idobot.SeedMsg) {}
	onEvent := func(_ idobot.Bot, _ *idobot.EventMsg) {}
	onError := func(_ idobot.Bot, _ error) {}
	bot, err := idobot.NewBot(url, apiToken, userAgent, "./test.db", onStart, onEvent, onError)
	defer bot.Stop()
	defer os.Remove("./test.db")

	if bot == nil || err != nil {
		t.Fatalf("idobot cannot instantiated with NewBot\n")
	}

	err = bot.PutDB("key", "value")
	if err != nil {
		t.Errorf("DBへの書き込みに失敗しました。\n")
	}

	v, err := bot.GetDB("key")
	if err != nil {
		t.Errorf("DBからの読み込みに失敗しました。\n")
	}

	if v != "value" {
		t.Errorf("DBから、想定していない値%sが読み込まれました。想定値は、%sです\n", v, "value")
	}
}
