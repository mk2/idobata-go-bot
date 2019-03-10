# What is this?

- Tool for making [idobata](https://idobata.io) bot by Go.

# How to use this?

1. import this library

```go
import "github.com/mk2/idobot"
```

2. initialize idobot

```go
var onStart = func(bot *idobot.Bot, msg *SeedMsg) {
 // do anything on initial message receiving.
}

var onEvent = func(bot *idobot.Bot, msg *EventMsg) {
 // do anything on each event message receiving.
}

// generate idobot by some settings.
bot, err := idobot.NewBot(idobataUrl, idobataApiToken, userAgent, onStart, onEvent)

// start bot
bot.Start()
```
