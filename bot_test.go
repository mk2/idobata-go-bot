package idobot_test

import (
	"testing"

	"github.com/mk2/idobot"
)

func TestIdobot_NewBot(t *testing.T) {
	url := "url"
	apiToken := "token"
	userAgent := "userAgent"
	onStart := func(_ idobot.Bot, _ *idobot.SeedMsg) {}
	onEvent := func(_ idobot.Bot, _ *idobot.EventMsg) {}
	bot, err := idobot.NewBot(url, apiToken, userAgent, onStart, onEvent)

	if bot == nil || err != nil {
		t.Errorf("idobot cannot instantiated with NewBot\n")
	}
}
