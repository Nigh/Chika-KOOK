package main

import (
	"regexp"
	"sync"
	"time"

	"github.com/lonelyevil/khl"
)

var jokes int = 0
var once sync.Once
var clockInput = make(chan interface{})
var commonChanSession *khl.Session

func commonChanHandlerInit(s *khl.Session) {
	commonChanSession = s
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

func unresponsiveMessageHandler(ctx *khl.TextMessageContext) {

	ctx.Session.MessageCreate(&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			TargetID: ctx.Common.TargetID,
			Content:  "这条是测试指令，只能由主人使用哦",
			Quote:    ctx.Common.MsgID,
		},
	})
}

func commonChanHandler(ctx *khl.TextMessageContext) {
	if ctx.Common.Type != khl.MessageTypeText {
		return
	}
	reply := func(words string) {
		ctx.Session.MessageCreate(&khl.MessageCreate{
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
	// r := regexp.MustCompile(`^\s*树脂\s*(\d{1,3})\s*$`)
	// matchs := r.FindStringSubmatch(ctx.Common.Content)
	// if len(matchs) > 0 {
	// 	count, _ := strconv.Atoi(matchs[1])

	// 	return
	// }
}
