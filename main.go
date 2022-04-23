package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"math/rand"

	kcard "local/khlcard"

	"github.com/jpillora/overseer"
	"github.com/jpillora/overseer/fetcher"
	"github.com/lonelyevil/khl"
	"github.com/lonelyevil/khl/log_adapter/plog"
	"github.com/phuslu/log"
	"github.com/spf13/viper"
)

// TODO:
// 仅保留masterID用于管理，上下线等调试信息直接私聊发送至master

var updateLog string = "增加删除账目功能"
var buildVersion string = "Chika-Zero Alpha0007"
var masterChannel string
var isVersionChange bool = false
var oneSession *khl.Session

func sendKCard(target string, content string) (resp *khl.MessageResp, err error) {
	return oneSession.MessageCreate((&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     khl.MessageTypeCard,
			TargetID: target,
			Content:  content,
		},
	}))
}
func sendMarkdown(target string, content string) (resp *khl.MessageResp, err error) {
	return oneSession.MessageCreate((&khl.MessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     khl.MessageTypeKMarkdown,
			TargetID: target,
			Content:  content,
		},
	}))
}

func sendMarkdownDirect(target string, content string) (mr *khl.MessageResp, err error) {
	return oneSession.DirectMessageCreate(&khl.DirectMessageCreate{
		MessageCreateBase: khl.MessageCreateBase{
			Type:     khl.MessageTypeKMarkdown,
			TargetID: target,
			Content:  content,
		},
	})
}

func prog(state overseer.State) {
	fmt.Printf("App#[%s] start ...\n", state.ID)
	rand.Seed(time.Now().UnixNano())
	viper.SetDefault("token", "0")
	viper.SetDefault("masterChannel", "0")
	viper.SetDefault("oldversion", "0.0.0")
	viper.SetDefault("lastwordsID", "")
	viper.SetConfigType("json")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	masterChannel = viper.Get("masterChannel").(string)

	if viper.Get("oldversion").(string) != buildVersion {
		isVersionChange = true
	}

	viper.Set("oldversion", buildVersion)

	l := log.Logger{
		Level:  log.InfoLevel,
		Writer: &log.ConsoleWriter{},
	}
	token := viper.Get("token").(string)
	fmt.Println("token=" + token)

	oneSession = khl.New(token, plog.NewLogger(&l))

	commonChanHandlerInit()
	accountBookInit()
	oneSession.AddHandler(messageHan)
	oneSession.AddHandler(markdownHan)
	oneSession.Open()

	if isVersionChange {
		go func() {
			<-time.After(time.Second * time.Duration(3))
			card := kcard.KHLCard{}
			card.Init()
			card.Card.Theme = "success"
			card.AddModule(
				kcard.KModule{
					Type: "header",
					Text: kcard.KField{
						Type:    "plain-text",
						Content: "Chika 热更新完成",
					},
				},
			)
			card.AddModule(
				kcard.KModule{
					Type: "divider",
				},
			)
			card.AddModule(
				kcard.KModule{
					Type: "section",
					Text: kcard.KField{
						Type:    "kmarkdown",
						Content: "当前版本号：`" + buildVersion + "`",
					},
				},
			)
			card.AddModule(
				kcard.KModule{
					Type: "section",
					Text: kcard.KField{
						Type:    "kmarkdown",
						Content: "**更新内容：**\n" + updateLog,
					},
				},
			)
			sendKCard(masterChannel, card.String())
		}()
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.")

	fmt.Println("[Read] lastwordsID=", viper.Get("lastwordsID").(string))
	if viper.Get("lastwordsID").(string) != "" {
		go func() {
			<-time.After(time.Second * time.Duration(7))
			oneSession.MessageDelete(viper.Get("lastwordsID").(string))
			viper.Set("lastwordsID", "")
		}()
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, overseer.SIGUSR2)
	<-sc

	lastResp, _ := sendMarkdown(masterChannel, randomSentence(shutdown))
	viper.Set("lastwordsID", lastResp.MsgID)
	fmt.Println("[Write] lastwordsID=", lastResp.MsgID)
	viper.WriteConfig()
	fmt.Println("Bot will shutdown after 1 second.")

	<-time.After(time.Second * time.Duration(1))
	// Cleanly close down the KHL session.
	oneSession.Close()
}

func main() {
	overseer.Run(overseer.Config{
		Required: true,
		Program:  prog,
		Fetcher:  &fetcher.File{Path: "Chika"},
		Debug:    false,
	})
}

func markdownHan(ctx *khl.KmarkdownMessageContext) {
	fmt.Printf("ctx.Common: %v\n", ctx.Common)
	fmt.Printf("ctx.Extra: %v\n", ctx.Extra)
}

func messageHan(ctx *khl.TextMessageContext) {
	fmt.Printf("ctx.Common: %v\n", ctx.Common)
	fmt.Printf("ctx.Extra: %v\n", ctx.Extra)
	if ctx.Extra.Author.Bot {
		return
	}
	commonChanHandler(ctx)
}
