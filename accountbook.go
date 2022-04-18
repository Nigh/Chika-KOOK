package main

import (
	"errors"
	"time"

	scribble "github.com/nanobox-io/golang-scribble"
)

// TODO:
// 若频道不可访问，则删除对应账本
// ~~只响应频道记账信息。~~
// 月度详细账单打包上传功能。
// 私聊只用于查询本人账单。

var db *scribble.Driver

var cacheRecord []accountRecord

type moneyRecord struct {
	User    string
	Time    int64
	Money   float64
	Comment string
	Id      string
}

type userRecord struct {
	User  string
	Money float64
}

type accountRecord struct {
	Id       string
	URecords []userRecord
	MRecords []moneyRecord
}

func allAccountSave() error {
	if err := db.Write("db", "allAccountRecord", cacheRecord); err != nil {
		return errors.New("账本保存失败")
	}
	return nil
}

func (a *accountRecord) Add(id string, user string, money float64, comment string) error {
	a.MRecords = append(a.MRecords, moneyRecord{user, time.Now().Unix(), money, comment, id})
	var found bool = false
	for i, v := range a.URecords {
		if v.User == user {
			a.URecords[i].Money += money
			found = true
			break
		}
	}
	if !found {
		a.URecords = append(a.URecords, userRecord{user, money})
	}
	return allAccountSave()
}

func (a *accountRecord) Tidy() {
	a.URecords = []userRecord{}
	for _, v := range a.MRecords {
		var found bool = false
		for i, u := range a.URecords {
			if v.User == u.User {
				found = true
				a.URecords[i].Money += v.Money
				break
			}
		}
		if !found {
			a.URecords = append(a.URecords, userRecord{v.User, v.Money})
		}
	}
}

func (a *accountRecord) Delete(id string, user string) error {
	for i := len(a.MRecords) - 1; i >= 0; i-- {
		if a.MRecords[i].Id == id {
			if a.MRecords[i].User == user {
				a.MRecords = append(a.MRecords[:i], a.MRecords[i+1:]...)
				a.Tidy()
				return allAccountSave()
			} else {
				return errors.New("不能删除非本人创建的账目")
			}
		}
	}
	return errors.New("未找到指定账目")
}

func accountBookInit() {
	db, _ = scribble.New("./database", nil)
	db.Read("db", "allAccountRecord", &cacheRecord)
	for i := range cacheRecord {
		cacheRecord[i].Tidy()
	}
}

func accountBookCreate(id string) error {
	for _, v := range cacheRecord {
		if v.Id == id {
			return errors.New("账本已经存在")
		}
	}
	cacheRecord = append(cacheRecord, accountRecord{id, []userRecord{}, []moneyRecord{}})
	return allAccountSave()
}

func accountBookGetSummary(id string) ([]userRecord, error) {
	for _, v := range cacheRecord {
		if v.Id == id {
			return v.URecords, nil
		}
	}
	return nil, errors.New("未找到账本")
}

func accountBookRecordAdd(groupId string, accountId string, user string, money float64, comment string) error {
	var found bool = false
	for i, v := range cacheRecord {
		if v.Id == groupId {
			cacheRecord[i].Add(accountId, user, money, comment)
			found = true
			break
		}
	}
	if !found {
		return errors.New("没有注册账本")
	}

	return allAccountSave()
}
func accountBookRecordDelete(groupId string, accountId string, user string) error {
	for i, v := range cacheRecord {
		if v.Id == groupId {
			err := cacheRecord[i].Delete(accountId, user)
			if err != nil {
				return err
			}
			return allAccountSave()
		}
	}
	return errors.New("未找到账本")
}
