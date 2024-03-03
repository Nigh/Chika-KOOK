package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	scribble "github.com/nanobox-io/golang-scribble"
)

// TODO:
// 月度详细账单打包上传功能。

type accountBook struct {
	// 所有账本
	Records map[string]*accountRecord
	// 有账本的群组ID
	Groups []groupRecord
}

type groupRecord struct {
	Id           string `json:"id"`
	AccountToken string `json:"token"`
}

type moneyRecord struct {
	User    string  `json:"user"`
	Time    int64   `json:"time"`
	Money   float64 `json:"money"`
	Comment string  `json:"comment"`
	Id      string  `json:"id"`
}

type userRecord struct {
	User  string  `json:"user"`
	Money float64 `json:"money"`
}
type userRecordList []userRecord

func (u *userRecordList) Push(user string, money float64) {
	for i, v := range *u {
		if v.User == user {
			(*u)[i].Money += money
			return
		}
	}
	*u = append(*u, userRecord{user, money})
}

type historyReport struct {
	User  string  `json:"user"`
	Money float64 `json:"money"` // 当月收支
	Total float64 `json:"total"` // 当月结算
}
type historyReportList []historyReport

func (h *historyReportList) Push(user string, money float64) {
	for i, v := range *h {
		if v.User == user {
			(*h)[i].Total += money
			(*h)[i].Money += money
			return
		}
	}
	*h = append(*h, historyReport{user, money, money})
}
func (h *historyReportList) ClearMoney() {
	for i := range *h {
		(*h)[i].Money = 0
	}
}

type historyRecord struct {
	Date   string            `json:"date"`
	Report historyReportList `json:"report"`
}

type historyDetail struct {
	User    string  `json:"user"`
	Time    string  `json:"time"`
	Money   float64 `json:"money"`
	Comment string  `json:"comment"`
}

type periodType string

const (
	ptHour  periodType = "hour"
	ptDay   periodType = "day"
	ptMonth periodType = "month"
)

type periodPay struct {
	// 余额
	Balance float64 `json:"balance"`
	// 注释
	Comment string `json:"comment"`
	// 付款金额
	Payment float64 `json:"pay"`
	// 周期类型
	PeriodType periodType `json:"type"`
	// 付款周期（小时/天/月）
	PayPeriod int `json:"period"`
	// 付款周期剩余
	PeriodLeft int `json:"nextPay"`
}
type periodPayList []periodPay

func (p *periodPayList) AddBalance(comment string, balance float64) error {
	for i, v := range *p {
		if v.Comment == comment {
			(*p)[i].Balance += balance
			return nil
		}
	}
	return errors.New("条目不存在")
}
func (p *periodPayList) SetBalance(comment string, balance float64) {
	for i, v := range *p {
		if v.Comment == comment {
			(*p)[i].Balance = balance
			return
		}
	}
	*p = append(*p, periodPay{balance, comment, 0, "", 0, 0})
}
func (p *periodPayList) SetPayment(comment string, pay float64, pt periodType, period int) error {
	for i, v := range *p {
		if v.Comment == comment {
			(*p)[i].Payment = pay
			(*p)[i].PeriodType = pt
			(*p)[i].PayPeriod = period
			(*p)[i].PeriodLeft = period
			return nil
		}
	}
	return errors.New("条目不存在")
}
func (p *periodPayList) Remove(comment string) error {
	for i, v := range *p {
		if v.Comment == comment {
			*p = append((*p)[:i], (*p)[i+1:]...)
			return nil
		}
	}
	return errors.New("条目不存在")
}

// return bad balance items
func (p *periodPayList) UpdateAtNewHour() []periodPay {
	ret := make([]periodPay, 0)
	for i, v := range *p {
		if v.PeriodType == ptHour ||
			(v.PeriodType == ptDay && time.Now().In(gTimezone).Hour() == 0) ||
			(v.PeriodType == ptMonth && time.Now().In(gTimezone).Day() == 1 && time.Now().In(gTimezone).Hour() == 0) {
			(*p)[i].PeriodLeft--
			if (*p)[i].PeriodLeft <= 0 {
				(*p)[i].Balance -= (*p)[i].Payment
				(*p)[i].PeriodLeft = (*p)[i].PayPeriod
				if (*p)[i].Balance < 0 {
					ret = append(ret, (*p)[i])
				}
			}
		}
	}
	return ret
}

type accountRecord struct {
	Id        string         `json:"id"` // 账本ID(群组ID)
	Token     string         `json:"token"`
	URecords  userRecordList `json:"users"`     // 当月基础数据
	PeriodPay periodPayList  `json:"periodpay"` // 周期付款
	MRecords  []moneyRecord  `json:"records"`   // 流水
}

func tokenGenerator() string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func (a *accountRecord) Add(id string, user string, money float64, comment string) error {
	a.MRecords = append(a.MRecords, moneyRecord{user, time.Now().Unix(), money, comment, id})
	exist := false
	for _, v := range a.URecords {
		if v.User == user {
			exist = true
		}
	}
	if !exist {
		a.URecords = append(a.URecords, userRecord{user, 0})
	}
	return a.Save()
}

func (a *accountRecord) Save() error {
	return db.Write(a.Id, "record", a)
}

func (a *accountRecord) Delete(id string, user string) error {
	for i := len(a.MRecords) - 1; i >= 0; i-- {
		if a.MRecords[i].Id == id {
			if a.MRecords[i].User == user {
				a.MRecords = append(a.MRecords[:i], a.MRecords[i+1:]...)
				return a.Save()
			} else {
				return errors.New("不能删除非本人创建的账目")
			}
		}
	}
	return errors.New("未找到指定账目")
}
func (a *accountRecord) Exist(id string) bool {
	for _, v := range a.MRecords {
		if v.Id == id {
			return true
		}
	}
	return false
}
func (a *accountRecord) GetComment(id string) string {
	for _, v := range a.MRecords {
		if v.Id == id {
			return v.Comment
		}
	}
	return "空"
}
func (a *accountRecord) GetToken() string {
	return a.Token
}
func (a *accountRecord) RefreshToken() {
	a.Token = tokenGenerator()
}

var db *scribble.Driver
var acout accountBook

func init() {
	db, _ = scribble.New("../database", nil)
	acout.Records = make(map[string]*accountRecord)
	db.Read("records", "groups", &acout.Groups)
	for _, v := range acout.Groups {
		var tmpRecord accountRecord
		db.Read(v.Id, "record", &tmpRecord)
		if tmpRecord.Token == "" {
			tmpRecord.Token = tokenGenerator()
		}
		acout.Records[v.Id] = &tmpRecord
	}
}

func (a *accountBook) Create(id string) error {
	if _, ok := a.Records[id]; ok {
		return errors.New("账本已经存在")
	}
	newToken := tokenGenerator()
	acout.Groups = append(acout.Groups, groupRecord{id, newToken})
	a.Records[id] = &accountRecord{id, newToken, userRecordList{}, periodPayList{}, []moneyRecord{}}
	db.Write("records", "groups", acout.Groups)
	return a.SaveById(id)
}
func (a *accountBook) SaveById(id string) error {
	if _, ok := a.Records[id]; !ok {
		return errors.New("账本不存在")
	}
	return db.Write(id, "record", a.Records[id])
}
func (a *accountBook) RemoveByChannel(id string) error {
	if _, ok := a.Records[id]; !ok {
		return errors.New("账本不存在")
	}
	for k := range a.Groups {
		if a.Groups[k].Id == id {
			history := make([]historyRecord, 0)
			db.Read(id, "history", &history)
			db.Write(id+"_bak", "history", history)
			db.Delete(id, "history")

			db.Write(id+"_bak", "record", a.Records[id])
			db.Delete(id, "record")

			delete(a.Records, id)
			a.Groups = append(a.Groups[:k], a.Groups[k+1:]...)
			return db.Write("records", "groups", a.Groups)
		}
	}
	return errors.New("未找到频道")
}

func (a *accountBook) SaveAll() error {
	db.Write("records", "groups", a.Groups)
	for _, v := range a.Groups {
		if err := a.SaveById(v.Id); err != nil {
			return errors.New("账本" + v.Id + "保存失败:" + err.Error())
		}
	}
	return nil
}

func (a *accountBook) Exist(groupId, accountId string) bool {
	if _, ok := a.Records[groupId]; !ok {
		return false
	}
	return a.Records[groupId].Exist(accountId)
}
func (a *accountBook) GetComment(groupId, msgId string) string {
	if _, ok := a.Records[groupId]; !ok {
		return "NULL"
	}
	return a.Records[groupId].GetComment(msgId)
}

func (a *accountBook) GetSummary(id string) (userRecordList, error) {
	if _, ok := a.Records[id]; !ok {
		return nil, errors.New("未找到账本")
	}
	ur := make(userRecordList, len(a.Records[id].URecords))
	copy(ur, a.Records[id].URecords)
	for _, v := range a.Records[id].MRecords {
		for i, u := range ur {
			if u.User == v.User {
				ur[i].Money += v.Money
			}
		}
	}
	return ur, nil
}

func (a *accountBook) RecordAdd(groupId string, recordId string, user string, money float64, comment string) error {
	if _, ok := a.Records[groupId]; !ok {
		return errors.New("没有注册账本")
	}
	return a.Records[groupId].Add(recordId, user, money, comment)
}

func (a *accountBook) RecordDelete(groupId string, accountId string, user string) error {
	if _, ok := a.Records[groupId]; !ok {
		return errors.New("未找到账本")
	}
	return a.Records[groupId].Delete(accountId, user)
}
