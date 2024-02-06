package main

import (
	"fmt"
	"regexp"
	"testing"
)

func TestDelete(t *testing.T) {
	r := regexp.MustCompile(`^删除\s+([0-9a-f\-]{16,48})`)
	matchs := r.FindStringSubmatch("删除 bca32308-a94a-4ee6-bc67-3753e1b9d86f")
	for _, v := range matchs {
		fmt.Println(v)
	}
}
