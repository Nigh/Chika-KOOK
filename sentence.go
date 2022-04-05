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

var resinAdded = []string{
	"你的期待将随时光流转而满盈。",
	"你的渴望将随日月流转而满足。",
	"你的欲望将随海潮涨落被填满。",
	"你的未来将随时针转动而到来。",
	"你的道路将随星辰大海而开辟。",
}

func randomSentence(list []string) string {
	return list[rand.Intn(len(list))]
}
