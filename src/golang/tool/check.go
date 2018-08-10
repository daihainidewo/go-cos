// Package tool tool
// file create by daihao, time is 2018/8/10 10:51
package tool

import (
	"os"
	"io/ioutil"
	"fmt"
	"golang/backup"
	"strings"
)

// CheckFileUpdate CheckFileUpdate
func IsFileUpdate(path string, ts int64) bool {
	// TODO
	mt, err := getFileUpdate(path)
	if err != nil {
		fmt.Println("get file update err", err)
		return true
	}
	if mt < ts {
		return false
	} else {
		return true
	}
}

// getFileUpdate getFileUpdate
func getFileUpdate(path string) (int64, error) {
	p, err := os.Stat(path)
	if err != nil {
		return -1, err
	}
	return p.ModTime().Unix(), nil
}

// IsDir IsDir
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// setChan setChan
func setChan(ch chan string, path string) error {
	if IsDir(path) {
		for _, d := range backup.Sysdata.Ig.GetIgdir() {
			i := strings.LastIndex(path, "/")
			if i < 0 {
				i = strings.LastIndex(path, "\\")
			}
			if d == path[i+1:]+"/" {
				return nil
			}
		}
		if path[len(path)-1] != '/' {
			path = path + "/"
		}

		files, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}
		for _, f := range files {
			setChan(ch, path+f.Name())
		}
	} else {
		for _, d := range backup.Sysdata.Ig.GetIgfile() {
			i := strings.LastIndex(path, "/")
			if i < 0 {
				i = strings.LastIndex(path, "\\")
			}
			if d == path[i+1:] {
				return nil
			}
		}
		ch <- path
	}
	return nil
}

// CheckDirUpdate CheckDirUpdate
func CheckDirUpdate(path string, ts int64) ([]string, error) {
	ret := make([]string, 0)
	name := make(chan string, 100)
	go func() {
		defer close(name)
		setChan(name, path)
	}()
	for {
		select {
		case n, ok := <-name:
			if !ok {
				return ret, nil
			}
			if IsFileUpdate(n, ts) {
				ret = append(ret, n)
			}
		default:
		}
	}
}
