// Package tool tool
// file create by daihao, time is 2018/8/14 20:40
package tool

import "runtime"

// PathLink PathLink
func PathLink(a, b string) string {
	return a + FileSymbol() + b
}

// FileSymbol 返回当前系统的文件分隔符
func FileSymbol() string {
	// TODO
	if runtime.GOOS == "windows" {
		return "\\"
	}
	return "/"
}
