package main

//import "fmt"
import (
	"math/rand"
)

var startUp = []string{
	"Chika 上线啦！",
	"世界第一可爱的 Chika 来啦！",
	"Chika 已唤醒。",
	"Chika，启动。",
}

var newVersionStartUp = []string{
	"内核升级完毕！ " + buildVersion + "型 Chika 上线啦！",
	"升级成功！全新的 " + buildVersion + "型 Chika 现在上线！",
	"Chika变得更可爱了！" + buildVersion + "型 Chika上线！",
}

var shutdown = []string{
	"啊，Chika 要充电了呢......",
	"啊，谁把 Chika 的电源线踢掉了...",
	"Chika 的电源线好像掉了...",
	"Chika 要换个新的模块了呢",
	"Chika 现在要去换件衣服哦",
	"Chika Chika Chika Chika Chika Chika",
}

func randomSentence(list []string) string {
	return list[rand.Intn(len(list))]
}
