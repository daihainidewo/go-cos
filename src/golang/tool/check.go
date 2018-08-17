// Package tool tool
// file create by daihao, time is 2018/8/10 10:51
package tool

import (
	"os"
	"io/ioutil"
	"fmt"
	"golang/sysfile"
	"strings"
	"sync"
)

// CheckFileUpdate CheckFileUpdate
func IsFileUpdate(path string, ts int64) bool {
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
func setChan(wg *sync.WaitGroup, ig *sysfile.Ignore, ch chan string, path string) {
	defer wg.Done()
	if IsDir(path) {
		for _, d := range ig.GetIgdir() {
			i := strings.LastIndex(path, FileSymbol())
			if d == path[i+1:]+"/" {
				return
			}
		}
		if path[len(path)-1] != '/' {
			path = path + "/"
		}

		files, err := ioutil.ReadDir(path)
		if err != nil {
			return
		}
		for _, f := range files {
			wg.Add(1)
			go setChan(wg, ig, ch, path+f.Name())
		}
	} else {
		for _, d := range ig.GetIgfile() {
			i := strings.LastIndex(path, FileSymbol())
			if d == path[i+1:] {
				return
			}
		}
		ch <- path
	}
}

// CheckDirUpdate CheckDirUpdate
func CheckDirUpdate(ig *sysfile.Ignore, path string, ts int64) ([]string, error) {
	ret := make([]string, 0)
	name := make(chan string, 100)
	go func() {
		defer close(name)
		wg := new(sync.WaitGroup)
		wg.Add(1)
		go setChan(wg, ig, name, path)
		wg.Wait()
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

// CheckDirUpdate2 CheckDirUpdate2
func CheckDirUpdate2(ig *sysfile.Ignore, path string, ts int64) ([]string, error) {
	// TODO
	ret := make([]string, 0)
	name := make(chan string, 100)
	go func() {
		defer close(name)
		wg := new(sync.WaitGroup)
		f := func(path string) {
			var ff func(path string) error
			ff = func(path string) error {
				defer wg.Done()
				if IsDir(path) {
					for _, d := range ig.GetIgdir() {
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
						wg.Add(1)
						go ff(path + f.Name())
					}
				} else {
					for _, d := range ig.GetIgfile() {
						i := strings.LastIndex(path, "/")
						if i < 0 {
							i = strings.LastIndex(path, "\\")
						}
						if d == path[i+1:] {
							return nil
						}
					}
					name <- path
				}
				return nil
			}
		}
		ff, f := f, nil
		wg.Add(1)
		go ff(path)
		wg.Wait()
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
