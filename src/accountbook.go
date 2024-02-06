package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	scribble "github.com/nanobox-io/golang-scribble"
)

// TODO:
// 若频道不可访问，则删除对应账本
// ~~只响应频道记账信息。~~
// 月度详细账单打包上传功能。
// 私聊只用于查询本人账单。

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
	Id       string // 账本ID(使用了群组ID)
	Token    string
	URecords []userRecord  // 用户余额
	MRecords []moneyRecord // 流水
}

func tokenGenerator() string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func allAccountSave() error {
	if err := db.Write("records", "currentMonth", cacheRecord); err != nil {
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

func init() {
	db, _ = scribble.New("../database", nil)
	// TODO: 每个群组的账本一个单独的目录
	db.Read("records", "currentMonth", &cacheRecord)
	for i := range cacheRecord {
		cacheRecord[i].Tidy()
		if cacheRecord[i].Token == "" {
			cacheRecord[i].Token = tokenGenerator()
		}
	}
}

func accountBookCreate(id string) error {
	for _, v := range cacheRecord {
		if v.Id == id {
			return errors.New("账本已经存在")
		}
	}
	cacheRecord = append(cacheRecord, accountRecord{id, tokenGenerator(), []userRecord{}, []moneyRecord{}})
	return allAccountSave()
}

func accountBookExist(id string) bool {
	for _, v := range cacheRecord {
		if v.Id == id {
			return true
		}
	}
	return false
}

func accountExist(groupId, accountId string) bool {
	for i, v := range cacheRecord {
		if v.Id == groupId {
			return cacheRecord[i].Exist(accountId)
		}
	}
	return false
}

func accountBookGet(groupId string) *accountRecord {
	for i, v := range cacheRecord {
		if v.Id == groupId {
			return &cacheRecord[i]
		}
	}
	return nil
}

func accountBookGetSummary(id string) ([]userRecord, error) {
	for _, v := range cacheRecord {
		if v.Id == id {
			return v.URecords, nil
		}
	}
	return nil, errors.New("未找到账本")
}

func accountBookRecordAdd(groupId string, recordId string, user string, money float64, comment string) error {
	var found bool = false
	for i, v := range cacheRecord {
		if v.Id == groupId {
			cacheRecord[i].Add(recordId, user, money, comment)
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
