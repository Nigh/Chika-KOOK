package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/lonelyevil/kook"
)

type handlerRule struct {
	matcher string
	getter  func(ctx *kook.EventHandlerCommonContext, matchs []string, reply func(string) string)
}

var commOnce sync.Once
var commRules []handlerRule = []handlerRule{
	{`^创建账本`, func(ctx *kook.EventHandlerCommonContext, s []string, f func(string) string) {
		err := acout.Create(ctx.Common.TargetID)
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

func accountCheck(ctx *kook.EventHandlerCommonContext, s []string, f func(string) string) {
	records, err := acout.GetSummary(ctx.Common.TargetID)
	if err != nil {
		f("错误:" + err.Error())
	} else {
		card := KookCard{}
		card.Init()
		card.AddModule(
			kkModule{
				Type: kkHeader,
				Text: kkField{
					Type:    kkPlaintext,
					Content: "总净支出",
				},
			},
		)
		card.AddModule(
			kkModule{
				Type: kkDivider,
			},
		)
		var userCol string = "**昵称**\n"
		var moneyCol string = "**净支出**\n"
		var diffCol string = "**净差额**\n"
		var min float64 = math.MaxFloat64
		for _, v := range records {
			if min > -v.Money {
				min = -v.Money
			}
			userCol += "(met)" + v.User + "(met)\n"
			moneyCol += strconv.FormatFloat(-v.Money, 'f', 2, 64) + "\n"
		}
		for _, v := range records {
			diffCol += "+" + strconv.FormatFloat(-v.Money-min, 'f', 2, 64) + "\n"
		}
		card.AddModule(
			kkModule{
				Type: kkSection,
				Text: kkField{
					Type: kkParagraph,
					Cols: 3,
					Fields: []kkField{
						{
							Type:    kkMarkdown,
							Content: userCol,
						},
						{
							Type:    kkMarkdown,
							Content: moneyCol,
						},
						{
							Type:    kkMarkdown,
							Content: diffCol,
						},
					},
				},
			},
		)
		sendKCard(ctx.Common.TargetID, card.String())
	}
}

func accountAdd(ctx *kook.EventHandlerCommonContext, s []string, f func(string) string) {
	money, _ := strconv.ParseFloat(s[2], 64)
	if s[1] == "支出" {
		money = -1 * money
	}
	comment := s[3]
	if len(comment) > 128 {
		comment = comment[:128] + "..."
	}

	err := acout.RecordAdd(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, money, comment)
	if err != nil {
		f("(met)" + ctx.Common.AuthorID + "(met) " + "错误:" + err.Error())
	} else {
		f("(met)" + ctx.Common.AuthorID + "(met) " + "记账成功，记账人点击记账下方的 ❌ 可以删除对应条目")
		oneSession.MessageAddReaction(ctx.Common.MsgID, "❌")
	}
}
func accountDelete(ctx *kook.EventHandlerCommonContext, s []string, f func(string) string) {
	err := acout.RecordDelete(ctx.Common.TargetID, s[1], ctx.Common.AuthorID)
	if err != nil {
		f("(met)" + ctx.Common.AuthorID + "(met) " + "错误:" + err.Error())
	} else {
		f("(met)" + ctx.Common.AuthorID + "(met) " + "账目已删除")
	}
}

func init() {
	commOnce.Do(func() { go clock() })
}

func clock() {
	min := time.NewTicker(1 * time.Minute)
	for range min.C {
		fmt.Println(time.Now().In(gTimezone).Format(gTimeFormat))
	}
}

func msgHandler(ctx *kook.KmarkdownMessageContext) {
	if ctx.Extra.Author.Bot {
		return
	}
	if ctx.Common.Type != kook.MessageTypeKMarkdown {
		return
	}
	reply := func(words string) string {
		resp, _ := sendMsg(ctx.Common.TargetID, words)
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

func reactionHandler(ctx *kook.ReactionAddContext) {
	u, _ := oneSession.UserMe()
	if ctx.Extra.UserID == u.ID {
		return
	}
	reply := func(words string) string {
		resp, _ := sendMsg(ctx.Extra.ChannelID, words)
		return resp.MsgID
	}
	go func() {
		if ctx.Extra.Emoji.ID == "❌" {
			if acout.Exist(ctx.Extra.ChannelID, ctx.Extra.MsgID) {
				comment := acout.GetComment(ctx.Extra.ChannelID, ctx.Extra.MsgID)
				err := acout.RecordDelete(ctx.Extra.ChannelID, ctx.Extra.MsgID, ctx.Extra.UserID)
				if err != nil {
					reply("(met)" + ctx.Extra.UserID + "(met) " + err.Error())
				} else {
					reply("(met)" + ctx.Extra.UserID + "(met) 成功删除了备注为`" + comment + "`的账目")
					oneSession.MessageDelete(ctx.Extra.MsgID)
				}
			}
		}
	}()
}
