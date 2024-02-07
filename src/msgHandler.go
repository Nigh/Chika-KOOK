package main

import (
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
	{`^余额\s+(\d+\.?\d*)\s*(.*)`, balanceSet},
	{`^查余额`, balanceList},
	{`^设置(\S+)\s+每(\d*)(天|月)扣款\s+(\d+\.?\d*)`, balancePaySet},
	{`^删除\s+([0-9a-f\-]{16,48})`, accountDelete},
}

func balanceSet(ctx *kook.EventHandlerCommonContext, s []string, f func(string) string) {
	pp := &acout.Records[ctx.Common.TargetID].PeriodPay
	n, _ := strconv.ParseFloat(s[1], 64)
	pp.SetBalance(s[2], n)
	f("已设置`" + s[2] + "`余额为 " + strconv.FormatFloat(n, 'f', 2, 64))
}
func balancePaySet(ctx *kook.EventHandlerCommonContext, s []string, f func(string) string) {
	pp := &acout.Records[ctx.Common.TargetID].PeriodPay
	n, _ := strconv.ParseFloat(s[4], 64)
	pd := 1
	var pt periodType
	if s[3] == "天" {
		pt = ptDay
	} else {
		pt = ptMonth
	}
	if len(s[2]) > 0 {
		pd, _ = strconv.Atoi(s[2])
	}
	err := pp.SetPayment(s[1], n, pt, pd)
	if err != nil {
		f(err.Error())
		return
	}
	f("已设置`" + s[1] + "`每" + strconv.Itoa(pd) + s[3] + "扣款 " + strconv.FormatFloat(n, 'f', 2, 64))
}

func balanceList(ctx *kook.EventHandlerCommonContext, s []string, f func(string) string) {
	pp := acout.Records[ctx.Common.TargetID].PeriodPay
	if len(pp) == 0 {
		f("无余额记录")
		return
	}
	card := KookCard{}
	card.Init()
	card.AddModule(
		kkModule{
			Type: kkHeader,
			Text: kkField{
				Type:    kkPlaintext,
				Content: "余额记录",
			},
		},
	)
	card.AddModule(
		kkModule{
			Type: kkDivider,
		},
	)
	var nameCol string = "**名称**\n"
	var paymentCol string = "**支付方式**\n"
	var balanceCol string = "**余额**\n"
	for _, v := range pp {
		nameCol += v.Comment + "\n"
		balanceCol += strconv.FormatFloat(v.Balance, 'f', 2, 64) + "\n"
		PeriodType := "每"
		if v.PayPeriod > 1 {
			PeriodType += strconv.Itoa(v.PayPeriod)
		}
		if v.PeriodType == ptDay {
			PeriodType += "天"
		}
		if v.PeriodType == ptMonth {
			PeriodType += "月"
		}
		PeriodType += "支付" + strconv.FormatFloat(v.Payment, 'f', 2, 64) + "\n"
		paymentCol += PeriodType
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
						Content: nameCol,
					},
					{
						Type:    kkMarkdown,
						Content: paymentCol,
					},
					{
						Type:    kkMarkdown,
						Content: balanceCol,
					},
				},
			},
		},
	)
	sendKCard(ctx.Common.TargetID, card.String())
}

func accountCheck(ctx *kook.EventHandlerCommonContext, s []string, f func(string) string) {
	records, err := acout.GetSummary(ctx.Common.TargetID)
	if err != nil {
		f("错误:" + err.Error())
		return
	}
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

func sendMonthSummary(groupId string, hr historyRecord) {
	card := KookCard{}
	card.Init()
	card.AddModule(
		kkModule{
			Type: kkHeader,
			Text: kkField{
				Type:    kkPlaintext,
				Content: hr.Date + " 月报",
			},
		},
	)
	card.AddModule(
		kkModule{
			Type: kkDivider,
		},
	)
	var userCol string = "**昵称**\n"
	var currentCol string = "**当月净支出**\n"
	var totalCol string = "**总净支出**\n"
	for _, v := range hr.Report {
		userCol += "(met)" + v.User + "(met)\n"
		currentCol += strconv.FormatFloat(-v.Money, 'f', 2, 64) + "\n"
		totalCol += strconv.FormatFloat(-v.Total, 'f', 2, 64) + "\n"
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
						Content: currentCol,
					},
					{
						Type:    kkMarkdown,
						Content: totalCol,
					},
				},
			},
		},
	)
	sendKCard(groupId, card.String())
}

func accountAdd(ctx *kook.EventHandlerCommonContext, s []string, f func(string) string) {
	money, _ := strconv.ParseFloat(s[2], 64)
	comment := s[3]
	rMoney := money
	rComment := comment
	if s[1] == "支出" {
		rMoney = -1 * rMoney
	}
	if len(rComment) > 128 {
		rComment = rComment[:128] + "..."
	}

	err := acout.RecordAdd(ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, rMoney, rComment)
	if err != nil {
		f("(met)" + ctx.Common.AuthorID + "(met) " + "错误:" + err.Error())
	} else {
		f("(met)" + ctx.Common.AuthorID + "(met) " + "记账成功，记账人点击记账下方的 ❌ 可以删除对应条目")
		oneSession.MessageAddReaction(ctx.Common.MsgID, "❌")
	}
	if acout.Records[ctx.Common.TargetID].PeriodPay.AddBalance(comment, money) == nil {
		f("成功为`" + comment + "`余额充值 " + strconv.FormatFloat(money, 'f', 2, 64))
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

func generateReport() {
	thisMonth := time.Now().In(gTimezone)
	previousMonth := thisMonth.AddDate(0, -1, 0)

	for _, v := range acout.Groups {
		history := make([]historyRecord, 0)
		db.Read(v.Id, "history", &history)
		for _, h := range history {
			if h.Date == previousMonth.Format("2006-01") {
				// 上月月报已经生成
				return
			}
		}
	}
	for _, v := range acout.Groups {
		// 建立月报
		monthReport := make(historyReportList, 0)

		// 复制累计数据
		for _, u := range acout.Records[v.Id].URecords {
			monthReport.Push(u.User, u.Money)
			monthReport.ClearMoney()
		}

		// 建立详细历史记录
		historyDetailList := make([]historyDetail, 0)
		// 保存非上月数据
		newMRecords := make([]moneyRecord, 0)

		// 统计上月数据
		for _, m := range acout.Records[v.Id].MRecords {
			// 过滤上月数据
			if time.Unix(m.Time, 0).In(gTimezone).Month() == previousMonth.Month() {
				// 月报数据
				monthReport.Push(m.User, m.Money)
				// 更新当前账本累计数据
				acout.Records[v.Id].URecords.Push(m.User, m.Money)
				// 详细历史记录
				historyDetailList = append(historyDetailList, historyDetail{
					User:    m.User,
					Time:    time.Unix(m.Time, 0).In(gTimezone).Format(gTimeFormat),
					Money:   m.Money,
					Comment: m.Comment,
				})
			} else {
				// 非上月数据保留
				newMRecords = append(newMRecords, m)
			}
		}

		// 持久化储存月报
		currentMonthRecord := historyRecord{
			Date:   previousMonth.Format("2006-01"),
			Report: monthReport,
		}
		history := make([]historyRecord, 0)
		db.Read(v.Id, "history", &history)
		history = append(history, currentMonthRecord)
		db.Write(v.Id, "history", history)

		// 持久化储存详细历史
		db.Write(v.Id, previousMonth.Format("2006-01"), historyDetailList)
		// 更新当前账本MRecords
		acout.Records[v.Id].MRecords = newMRecords
		// 持久化储存当前账本
		acout.SaveById(v.Id)

		// 发送月报，3秒间隔，防止机器人被Ban
		sendMonthSummary(v.Id, currentMonthRecord)
		<-time.After(3 * time.Second)
	}
}

func clock() {
	tick := time.NewTicker(15 * time.Minute)
	done := false
	newday := false
	for range tick.C {
		if !done && time.Now().In(gTimezone).Day() == 1 && time.Now().In(gTimezone).Hour() >= 4 {
			generateReport()
			done = true
		}
		if time.Now().In(gTimezone).Day() != 1 {
			done = false
		}
		if !newday && time.Now().In(gTimezone).Hour() == 0 {
			newday = true
			for k := range acout.Records {
				acout.Records[k].PeriodPay.UpdateAtNewDay()
				bb := acout.Records[k].PeriodPay.GetBadBalanceItem()
				if len(bb) > 0 {
					for _, v := range bb {
						sendMsg(k, "`"+v.Comment+"` 余额已不足，请及时充值！")
					}
				}
			}
		}
		if time.Now().In(gTimezone).Hour() == 23 {
			newday = false
		}
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
