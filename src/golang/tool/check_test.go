// Package tool tool
// file create by daihao, time is 2018/8/10 11:30
package tool

import (
	"testing"
	"fmt"
	"golang/sysfile"
)

func TestCheckDirUpdate(t *testing.T) {
	sysfile.Sysdata = sysfile.NewSysData("/home/daihao/git/go-cos/.gocos")
	err := sysfile.Sysdata.Ig.Read("/home/daihao/git/go-cos/.gocos/ignore")
	if err != nil {
		fmt.Println(err)
	}
	res, err := CheckDirUpdate("/home/daihao/git/go-cos", 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(res))
	for _, d := range res {
		fmt.Println(d)
	}
}
