package main

import (
	"testing"
)

func TestInit(t *testing.T) {
	accountBookInit()
}

func TestAB(t *testing.T) {
	accountBookInit()
	accountBookCreate("2980776678707937")
	accountBookRecordAdd("2980776678707937", "3797446169", 100, "")
	accountBookRecordAdd("2980776678707937", "3797446169", -100, "")
	accountBookRecordAdd("2980776678707937", "a3797446169", 100, "")
	accountBookRecordAdd("12345", "a3797446169", 100, "")
	accountBookCreate("12345")
	accountBookRecordAdd("12345", "a3797446169", 100, "")
	accountBookCreate("12345")
}
