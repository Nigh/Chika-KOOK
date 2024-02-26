package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	kuma "github.com/Nigh/kuma-push"
	"github.com/lonelyevil/kook"
	"github.com/lonelyevil/kook/log_adapter/plog"
	"github.com/phuslu/log"
)

var (
	gTimezone    *time.Location = time.FixedZone("CST", 8*60*60)
	gTimeFormat  string         = "2006-01-02 15:04"
	gKumaPushURL string
	gToken       string
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
	gToken = os.Getenv("BOT_TOKEN")
	gKumaPushURL = os.Getenv("KUMA_PUSH_URL")
	fmt.Println("token=" + gToken)
	if gToken == "" {
		fmt.Println("Bot token not set.")
		<-time.After(time.Second * 3)
		os.Exit(1)
	}

	l := log.Logger{
		Level:  log.InfoLevel,
		Writer: &log.ConsoleWriter{},
	}

	oneSession = kook.New(gToken, plog.NewLogger(&l))
	oneSession.AddHandler(msgHandler)
	oneSession.AddHandler(reactionHandler)
	oneSession.Open()

	k := kuma.New(gKumaPushURL)
	k.Start()
	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.")

	go func() {
		defer func() {
			fmt.Println("Bot is restarting.")
			<-time.After(time.Second * 1)
			oneSession.Close()
			os.Exit(2)
		}()
		minute := time.NewTicker(1 * time.Minute)
		restartTimer := 1440
		for range minute.C {
			restartTimer -= 1
			if time.Now().Minute() == 0 && time.Now().Hour() == 5 {
				return
			}
			if time.Since(oneSession.LastHeartbeatAck).Seconds() > 600 || restartTimer <= 0 {
				return
			}
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)
	<-sc
	err := acout.SaveAll()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("Bot will shutdown after 1 second.")

	<-time.After(time.Second * 1)
	oneSession.Close()
}
