package main

import (
	"regexp"
	"sync"
	"time"

	"github.com/lonelyevil/khl"
)

var once sync.Once
var clockInput = make(chan interface{})

func commonChanHandlerInit() {
	once.Do(func() { go clock(clockInput) })
}

func clock(input chan interface{}) {
	min := time.NewTicker(1 * time.Minute)
	halfhour := time.NewTicker(23 * time.Minute)
	for {
		select {
		case <-min.C:
			hour, min, _ := time.Now().Local().Clock()
			if min == 0 && hour == 5 {
			}
		case <-halfhour.C:
		}
	}
}

func commonChanHandler(ctx *khl.TextMessageContext) {
	handlerSession := ctx.Session
	if ctx.Common.Type != khl.MessageTypeText {
		return
	}
	reply := func(words string) {
		handlerSession.MessageCreate(&khl.MessageCreate{
			MessageCreateBase: khl.MessageCreateBase{
				Type:     khl.MessageTypeKMarkdown,
				TargetID: masterChannel,
				Content:  words,
			},
		})
	}
	match, _ := regexp.MatchString("^Chika在么.{0,5}", ctx.Common.Content)
	if match {
		reply("Chika在的哦")
		return
	}
}
