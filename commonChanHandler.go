package main

import (
	kcard "local/khlcard"
	"regexp"
	"strconv"
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
				TargetID: ctx.Common.TargetID,
				Content:  words,
			},
		})
	}
	match, _ := regexp.MatchString("^Chika在么.{0,5}", ctx.Common.Content)
	if match {
		reply("Chika在的哦")
		return
	}
	match, _ = regexp.MatchString("^创建账本", ctx.Common.Content)
	if match {
		err := accountBookCreate(ctx.Common.TargetID)
		if err != nil {
			reply("错误:" + err.Error())
		} else {
			reply("账本已创建")
		}
		return
	}
	r := regexp.MustCompile(`^(支出|收入)\s+(\d+\.?\d*)\s*(.*)`)
	matchs := r.FindStringSubmatch(ctx.Common.Content)
	if len(matchs) > 0 {
		money, _ := strconv.ParseFloat(matchs[2], 64)
		if matchs[1] == "支出" {
			money = -1 * money
		}
		comment := matchs[3]
		user := ctx.Common.AuthorID
		err := accountBookRecordAdd(ctx.Common.TargetID, user, money, comment)
		if err != nil {
			reply("错误:" + err.Error())
		} else {
			reply("记账成功")
		}
		return
	}
	match, _ = regexp.MatchString("^查账", ctx.Common.Content)
	records, err := accountBookGetSummary(ctx.Common.TargetID)
	if err != nil {
		reply("错误:" + err.Error())
	} else {
		card := kcard.KHLCard{}
		card.Init()
		card.AddModule(
			kcard.KModule{
				Type: kcard.Header,
				Text: kcard.KField{
					Type:    kcard.Plaintext,
					Content: "总净资产",
				},
			},
		)
		card.AddModule(
			kcard.KModule{
				Type: kcard.Divider,
			},
		)
		var userCol string = "**昵称**\n"
		var moneyCol string = "**净资产**\n"
		for _, v := range records {
			userCol += "(met)" + v.User + "(met)\n"
			moneyCol += strconv.FormatFloat(v.Money, 'f', 2, 64) + "\n"
		}
		card.AddModule(
			kcard.KModule{
				Type: kcard.Section,
				Text: kcard.KField{
					Type: kcard.Paragraph,
					Cols: 2,
					Fields: []kcard.KField{
						{
							Type:    kcard.Kmarkdown,
							Content: userCol,
						},
						{
							Type:    kcard.Kmarkdown,
							Content: moneyCol,
						},
					},
				},
			},
		)
	}
	return
}
