package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lonelyevil/kook"
	"github.com/lonelyevil/kook/log_adapter/plog"
	"github.com/phuslu/log"
)

var (
	gTimezone   *time.Location = time.FixedZone("CST", 8*60*60)
	gTimeFormat string         = "2006-01-02 15:04"
)

var oneSession *kook.Session

func sendKCard(target string, content string) (resp *kook.MessageResp, err error) {
	return oneSession.MessageCreate((&kook.MessageCreate{
		MessageCreateBase: kook.MessageCreateBase{
			Type:     kook.MessageTypeCard,
			TargetID: target,
			Content:  content,
		},
	}))
}
func sendMsg(target string, content string) (resp *kook.MessageResp, err error) {
	return oneSession.MessageCreate((&kook.MessageCreate{
		MessageCreateBase: kook.MessageCreateBase{
			Type:     kook.MessageTypeKMarkdown,
			TargetID: target,
			Content:  content,
		},
	}))
}

func sendMsgDirect(target string, content string) (mr *kook.MessageResp, err error) {
	return oneSession.DirectMessageCreate(&kook.DirectMessageCreate{
		MessageCreateBase: kook.MessageCreateBase{
			Type:     kook.MessageTypeKMarkdown,
			TargetID: target,
			Content:  content,
		},
	})
}

func main() {
	token := os.Getenv("BOT_TOKEN")
	fmt.Println("token=" + token)
	if token == "" {
		fmt.Println("Bot token not set.")
		<-time.After(time.Second * 3)
		os.Exit(1)
	}

	l := log.Logger{
		Level:  log.InfoLevel,
		Writer: &log.ConsoleWriter{},
	}

	oneSession = kook.New(token, plog.NewLogger(&l))
	oneSession.AddHandler(msgHandler)
	oneSession.AddHandler(reactionHandler)
	oneSession.Open()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	<-sc

	fmt.Println("Bot will shutdown after 1 second.")

	<-time.After(time.Second * 1)
	oneSession.Close()
}
