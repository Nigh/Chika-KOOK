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
	{`^帮助$`, help},
	{`^创建账本`, accountCreate},
	{`^查账`, accountCheck},
	{`^查自动扣款`, balanceList},
	{`^(支出|收入)\s+(\d+\.?\d*)\s*(\S+)`, accountAdd},
	{`^转账\s+(\d+\.?\d*)\s*\(met\)(\d+)\(met\)`, transferRequest},
	{`^设置余额\s+(\d+\.?\d*)\s*(\S+)`, balanceSet},
	{`^删除余额\s+(\S+)`, balanceRemove},
	{`^设置(\S+)\s+每(\d*)(小时|天|月)扣款\s+(\d+\.?\d*)`, balancePaySet},
	{`^删除\s+([0-9a-f\-]{16,48})`, accountDelete},
}

// TODO: 帮助卡片
func help(ctx *kook.EventHandlerCommonContext, s []string, f func(string) string) {
	helpStr := "命令列表:\n"
	helpStr += "`帮助`：显示本条帮助\n"
	helpStr += "`创建账本`：为频道创建账本，每个频道只能创建一个账本\n"
	helpStr += "`查账`：查看总支出\n"
	helpStr += "`查自动扣款`：查看自动扣款余额\n"
	helpStr += "`支出 金额 备注` 或 `收入 金额 备注`：添加一条账本记录\n"
	helpStr += "`转账 金额 @某用户`：发起一个转账请求\n"
	helpStr += "`设置余额 金额 备注`：设置一个自动扣款记录的余额，没有则会新建\n"
	helpStr += "`删除余额 备注`：删除一个自动扣款记录\n"
	helpStr += "`设置备注 每n[小时/天/月]扣款 金额`：设置一个自动扣款记录的扣款方式\n\n"
	helpStr += "示例:\n"
	helpStr += "`支出 1000 交通费`\n"
	helpStr += "`设置余额 800 停车费`\n"
	helpStr += "`设置停车费 每天扣款 40`\n"
	helpStr += "`设置余额 128000 停机坪`\n"
	helpStr += "`设置停机坪 每6小时扣款 800`\n"
	helpStr += "`支出 2000 停机坪`\n\n"
	helpStr += "当自动扣款余额不足时，将会发布消息提醒。\n当支出项与自动扣款备注相同时，会自动将支出金额添加到自动扣款的余额中。"
	f(helpStr)
}
func balanceRemove(ctx *kook.EventHandlerCommonContext, s []string, f func(string) string) {
	pp := &acout.Records[ctx.Common.TargetID].PeriodPay
	err := pp.Remove(s[1])
	if err != nil {
		f(err.Error())
	} else {
		f("已删除`" + s[1] + "`的自动扣款记录")
	}
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
	switch s[3] {
	case "小时":
		pt = ptHour
	case "天":
		pt = ptDay
	case "月":
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
	var paymentCol string = "**扣款方式**\n"
	var balanceCol string = "**当前余额**\n"
	for _, v := range pp {
		nameCol += v.Comment + "\n"
		balanceCol += strconv.FormatFloat(v.Balance, 'f', 2, 64) + "\n"
		PeriodType := "每"
		if v.PayPeriod > 1 {
			PeriodType += strconv.Itoa(v.PayPeriod)
		}
		if v.PeriodType == ptHour {
			PeriodType += "小时"
		}
		if v.PeriodType == ptDay {
			PeriodType += "天"
		}
		if v.PeriodType == ptMonth {
			PeriodType += "月"
		}
		PeriodType += "扣款" + strconv.FormatFloat(v.Payment, 'f', 2, 64) + "\n"
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
		userCol += userAt(v.User) + "\n"
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
		userCol += userAt(v.User) + "\n"
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

func accountCreate(ctx *kook.EventHandlerCommonContext, s []string, f func(string) string) {
	err := acout.Create(ctx.Common.TargetID)
	if err != nil {
		f(userAt(ctx.Common.AuthorID) + " 错误:" + err.Error())
	} else {
		f(userAt(ctx.Common.AuthorID) + " 账本已创建")
	}
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
		f(userAt(ctx.Common.AuthorID) + " 错误:" + err.Error())
	} else {
		f(userAt(ctx.Common.AuthorID) + " 记账成功，记账人点击记账下方的 ❌ 可以删除对应条目")
		oneSession.MessageAddReaction(ctx.Common.MsgID, "❌")
	}
	if acout.Records[ctx.Common.TargetID].PeriodPay.AddBalance(comment, money) == nil {
		f("成功为`" + comment + "`余额充值 " + strconv.FormatFloat(money, 'f', 2, 64))
	}
}

type transferPending struct {
	channelID string
	msgID     string
	fromID    string
	toID      string
	money     float64
	timeLeft  int
}

var transferPendingList []transferPending

func transferRequest(ctx *kook.EventHandlerCommonContext, s []string, f func(string) string) {
	userID := s[2]
	money, _ := strconv.ParseFloat(s[1], 64)
	f(userAt(ctx.Common.AuthorID) + " 向 " + userAt(userID) + " 的转账请求已发起，收款方点击转账请求下方的 ✅ 即表示已经收款\n十分钟内未完成的转账将会被自动取消")
	transferPendingList = append(transferPendingList, transferPending{ctx.Common.TargetID, ctx.Common.MsgID, ctx.Common.AuthorID, userID, money, 60})
	oneSession.MessageAddReaction(ctx.Common.MsgID, "✅")
}
func accountDelete(ctx *kook.EventHandlerCommonContext, s []string, f func(string) string) {
	err := acout.RecordDelete(ctx.Common.TargetID, s[1], ctx.Common.AuthorID)
	if err != nil {
		f(userAt(ctx.Common.AuthorID) + " 错误:" + err.Error())
	} else {
		f(userAt(ctx.Common.AuthorID) + " 账目已删除")
	}
}

func init() {
	commOnce.Do(func() { go clock() })
	commOnce.Do(func() { go transferTimer() })
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

		// 发送月报，添加间隔，防止机器人被Ban
		fmt.Printf("Send Month Summary to Channel: [%s]\n", v.Id)
		sendMonthSummary(v.Id, currentMonthRecord)
		<-time.After(1 * time.Second)
	}
}
func transferTimer() {
	tick := time.NewTicker(10 * time.Second)
	for range tick.C {
		newList := make([]transferPending, 0)
		for idx := range transferPendingList {
			transferPendingList[idx].timeLeft--
			if transferPendingList[idx].timeLeft > 0 {
				newList = append(newList, transferPendingList[idx])
			} else {
				sendMsg(transferPendingList[idx].channelID, fmt.Sprintf("**注意**：%s -> %s 的转账请求已超时失效\n", userAt(transferPendingList[idx].fromID), userAt(transferPendingList[idx].toID)))
				oneSession.MessageDelete(transferPendingList[idx].msgID)
			}
		}
		transferPendingList = newList
	}
}
func clock() {
	invalidChannels := make(map[string]int, 0)
	tick := time.NewTicker(17 * time.Minute)
	lastUpdateHour := time.Now().In(gTimezone).Hour()
	for range tick.C {
		if lastUpdateHour != time.Now().In(gTimezone).Hour() {
			fmt.Printf("New Hour Tick\n")
			lastUpdateHour = time.Now().In(gTimezone).Hour()
			if time.Now().In(gTimezone).Day() == 1 && time.Now().In(gTimezone).Hour() == 4 {
				generateReport()
			}
			for k := range acout.Records {
				_, err := oneSession.ChannelView(k)
				if err != nil {
					// 频道无法访问
					fmt.Printf("Channel[%s] access error: %s\n", k, err.Error())
					if _, ok := invalidChannels[k]; !ok {
						invalidChannels[k] = 1
					} else {
						invalidChannels[k]++
						fmt.Printf("Channel[%s] invalid count: %d\n", k, invalidChannels[k])
						if invalidChannels[k] > 11 {
							// 连续超过11次频道无法访问，认为频道已经移除
							delete(invalidChannels, k)
							acout.RemoveByChannel(k)
							fmt.Printf("Channel[%s] removed\n", k)
						}
					}
					continue
				} else {
					invalidChannels[k] = 0
				}
				acout.Records[k].PeriodPay.UpdateAtNewHour()
				bb := acout.Records[k].PeriodPay.GetBadBalanceItem()
				if len(bb) > 0 {
					for _, v := range bb {
						sendMsg(k, "`"+v.Comment+"` 余额已不足，请及时充值！")
					}
				}
			}
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
		switch ctx.Extra.Emoji.ID {
		case "❌":
			if acout.Exist(ctx.Extra.ChannelID, ctx.Extra.MsgID) {
				comment := acout.GetComment(ctx.Extra.ChannelID, ctx.Extra.MsgID)
				err := acout.RecordDelete(ctx.Extra.ChannelID, ctx.Extra.MsgID, ctx.Extra.UserID)
				if err != nil {
					reply(userAt(ctx.Extra.UserID) + " " + err.Error())
				} else {
					reply(userAt(ctx.Extra.UserID) + " 成功删除了备注为`" + comment + "`的账目")
					oneSession.MessageDelete(ctx.Extra.MsgID)
				}
			}
		case "✅":
			for i, v := range transferPendingList {
				if v.channelID == ctx.Extra.ChannelID && v.msgID == ctx.Extra.MsgID {
					if v.toID != ctx.Extra.UserID {
						reply(userAt(ctx.Extra.UserID) + " 你不能确认别人的转账")
						break
					}
					err := acout.RecordAdd(v.channelID, v.msgID, v.fromID, -v.money, "转账给"+userAt(v.toID))
					acout.RecordAdd(v.channelID, v.msgID, v.toID, v.money, "接受"+userAt(v.fromID)+"的转账")
					if err != nil {
						reply(userAt(ctx.Extra.UserID) + " 错误:" + err.Error())
					} else {
						reply(userAt(ctx.Extra.UserID) + " 您已成功确认" + userAt(v.fromID) + "的转账")
					}
					transferPendingList = append(transferPendingList[:i], transferPendingList[i+1:]...)
					break
				}
			}
		}
	}()
}

func userAt(id string) string {
	return "(met)" + id + "(met)"
}
