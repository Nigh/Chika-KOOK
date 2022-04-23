package main

import (
	kcard "local/khlcard"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/lonelyevil/khl"
)

type handlerRule struct {
	matcher string
	getter  func(ctx *khl.EventHandlerCommonContext, matchs []string, reply func(string) string)
}

var commOnce sync.Once
var clockInput = make(chan interface{})
var commRules []handlerRule = []handlerRule{
	{`^Chika在么.{0,5}`, func(ctx *khl.EventHandlerCommonContext, s []string, f func(string) string) {
		msgId := f("Chika在的哦")
		go func(id []string) {
			<-time.After(time.Second * time.Duration(5))
			for _, v := range id {
				oneSession.MessageDelete(v)
			}
		}([]string{msgId, ctx.Common.MsgID})
	}},
	{`^创建账本`, func(ctx *khl.EventHandlerCommonContext, s []string, f func(string) string) {
		err := accountBookCreate(ctx.Common.TargetID)
		if err != nil {
			f("(met)" + ctx.Common.AuthorID + "(met) " + "错误:" + err.Error())
		} else {
			f("(met)" + ctx.Common.AuthorID + "(met) " + "账本已创建")
		}
	}},
	{`^查账`, accountCheck},
	{`^(支出|收入)\s+(\d+\.?\d*)\s*(.*)`, accountAdd},
	{`^删除\s+([0-9a-f\-]{16,48})`, accountDelete},
}

func accountCheck(ctx *khl.EventHandlerCommonContext, s []string, f func(string) string) {
	records, err := accountBookGetSummary(ctx.Common.TargetID)
	if err != nil {
		f("错误:" + err.Error())
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
		sendKCard(ctx.Common.TargetID, card.String())
	}
}

func accountAdd(ctx *khl.EventHandlerCommonContext, s []string, f func(string) string) {
	money, _ := strconv.ParseFloat(s[2], 64)
	if s[1] == "支出" {
		money = -1 * money
	}
	comment := s[3]
	err := accountBookRecordAdd(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, money, comment)
	if err != nil {
		f("(met)" + ctx.Common.AuthorID + "(met) " + "错误:" + err.Error())
	} else {
		f("(met)" + ctx.Common.AuthorID + "(met) " + "记账成功，账目ID:`" + ctx.Common.MsgID + "`")
	}
}
func accountDelete(ctx *khl.EventHandlerCommonContext, s []string, f func(string) string) {
	err := accountBookRecordDelete(ctx.Common.TargetID, s[1], ctx.Common.AuthorID)
	if err != nil {
		f("(met)" + ctx.Common.AuthorID + "(met) " + "错误:" + err.Error())
	} else {
		f("(met)" + ctx.Common.AuthorID + "(met) " + "账目已删除")
	}
}

func commonChanHandlerInit() {
	commOnce.Do(func() { go clock(clockInput) })
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

func commonChanMarkdownHandler(ctx *khl.KmarkdownMessageContext) {
	if ctx.Common.Type != khl.MessageTypeText && ctx.Common.Type != khl.MessageTypeKMarkdown {
		return
	}
	reply := func(words string) string {
		resp, _ := sendMarkdown(ctx.Common.TargetID, words)
		return resp.MsgID
	}
	for n := range commRules {
		v := &commRules[n]
		r := regexp.MustCompile(v.matcher)
		matchs := r.FindStringSubmatch(ctx.Common.Content)
		if len(matchs) > 0 {
			go v.getter(ctx.EventHandlerCommonContext, matchs, reply)
			return
		}
	}
}

func commonChanHandler(ctx *khl.TextMessageContext) {
	if ctx.Common.Type != khl.MessageTypeText && ctx.Common.Type != khl.MessageTypeKMarkdown {
		return
	}
	reply := func(words string) string {
		resp, _ := sendMarkdown(ctx.Common.TargetID, words)
		return resp.MsgID
	}
	for n := range commRules {
		v := &commRules[n]
		r := regexp.MustCompile(v.matcher)
		matchs := r.FindStringSubmatch(ctx.Common.Content)
		if len(matchs) > 0 {
			go v.getter(ctx.EventHandlerCommonContext, matchs, reply)
			return
		}
	}
}
