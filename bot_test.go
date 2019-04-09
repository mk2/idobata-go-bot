package idobot_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mk2/idobot"
)

func TestIdobot_NewBot(t *testing.T) {
	url := "url"
	apiToken := "token"
	userAgent := "userAgent"
	onStart := func(_ idobot.Bot, _ *idobot.SeedMsg) {}
	onEvent := func(_ idobot.Bot, _ *idobot.EventMsg) {}
	onError := func(_ idobot.Bot, _ error) {}
	bot, err := idobot.NewBot(url, apiToken, userAgent, onStart, onEvent, onError)

	if bot == nil || err != nil {
		t.Errorf("idobot cannot instantiated with NewBot\n")
	}
}

func TestIdobot_PostMessage(t *testing.T) {
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
	bot, err := idobot.NewBot(url, apiToken, userAgent, onStart, onEvent, onError)

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
