// Package main main
// file create by daihao, time is 2018/8/30 11:45
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/gommon/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Ignore
type Ignore struct {
	dir  map[string]struct{} // 结尾不带分割符
	file map[string]struct{}
}

// NewIgnore new Ignore
func NewIgnore(path string) *Ignore {
	ret := &Ignore{
		dir:  make(map[string]struct{}, 0),
		file: make(map[string]struct{}, 0),
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("read ignore file error:", err)
		return nil
	}
	res := bytes.Split(b, []byte("\n"))
	for _, d := range res {
		d = bytes.TrimSpace(d)
		if len(d) == 0 {
			continue
		}
		if d[len(d)-1] == '/' {
			ret.AddIgdir(string(d[:len(d)-1]))
		} else {
			ret.AddIgfile(string(d))
		}
	}
	return ret
}

// AddIgdir AddIgdir
func (ig *Ignore) AddIgdir(dirname string) {
	ig.dir[dirname] = struct{}{}
}

// AddIgfile AddIgfile
func (ig *Ignore) AddIgfile(filename string) {
	ig.file[filename] = struct{}{}
}

// GetIgdir GetIgdir
func (ig *Ignore) GetIgdir() map[string]struct{} {
	return ig.dir
}

// GetIgfile GetIgfile
func (ig *Ignore) GetIgfile() map[string]struct{} {
	return ig.file
}

// FileNodeSlice
type FileNodeSlice struct {
	Slice    []string `json:"slice"` // 相对于rootpath的相对路径
	sm       *sync.Mutex
	ig       *Ignore
	rootPath string
}

// NewFileNodeSlice new FileNodeSlice
func NewFileNodeSlice(rootpath string, ig *Ignore) *FileNodeSlice {
	return &FileNodeSlice{
		Slice:    make([]string, 0, 10),
		sm:       new(sync.Mutex),
		ig:       ig,
		rootPath: rootpath,
	}
}

// String String
func (fns *FileNodeSlice) String() (string) {
	ret := "paths:\n"
	for _, d := range fns.Slice {
		ret += d + "\n"
	}
	return ret
}

// Len Len
func (fns *FileNodeSlice) Len() (int) {
	return len(fns.Slice)
}

// Less Less
func (fns *FileNodeSlice) Less(i, j int) (bool) {
	return fns.Slice[i] < fns.Slice[j]
}

// Swap Swap
func (fns *FileNodeSlice) Swap(i, j int) () {
	fns.Slice[i], fns.Slice[j] = fns.Slice[j], fns.Slice[i]
}

// Check 查出新建文件，修改文件，删除文件, last必须是sort后的相对路径序列，timestamp是上次修改的时间戳，返回路径为绝对路径
func (fns *FileNodeSlice) Check(last []string, timestamp int64) (crt []string, mod []string, del []string) {
	// TODO
	a, b := 0, 0
	crt = make([]string, 0)
	mod = make([]string, 0)
	del = make([]string, 0)
	for a < len(fns.Slice) && b < len(last) {
		if fns.Slice[a] == last[b] {
			ap := filepath.Join(syscos.rootpath, fns.Slice[a])
			info, err := os.Lstat(ap)
			if err != nil {
				log.Errorf("os lstat %s, error:%s", ap, err)
				a++
				b++
				continue
			}
			if info.ModTime().Unix() > timestamp {
				mod = append(mod, fns.Slice[a])
			}
			a++
			b++
		} else if fns.Slice[a] > last[b] {
			del = append(del, last[b])
			b++
		} else {
			crt = append(crt, fns.Slice[a])
			a++
		}
	}

	if a == len(fns.Slice) {
		del = append(del, last[b:]...)
	} else {
		crt = append(crt, fns.Slice[a:]...)
	}

	return
}

// Sort Sort
func (fns *FileNodeSlice) Sort() () {
	sort.Sort(fns)
}

// Add Add
func (fns *FileNodeSlice) Add(n string) () {
	fns.sm.Lock()
	defer fns.sm.Unlock()
	fns.Slice = append(fns.Slice, n)
}

// Adds Adds
func (fns *FileNodeSlice) Adds(n ...string) () {
	fns.sm.Lock()
	defer fns.sm.Unlock()
	fns.Slice = append(fns.Slice, n...)
}

// ToStringArray ToStringArray
func (fns *FileNodeSlice) ToStringArray() ([]string) {
	return fns.Slice
}

// Walk Walk
func (fns *FileNodeSlice) Walk(path string) ([]string) {
	wg := new(sync.WaitGroup)
	token := NewTokenBucket(1000)
	token.Get()
	wg.Add(1)
	go fns.w(path, wg, &token)
	wg.Wait()
	fns.Sort()
	return fns.ToStringArray()
}

// Write Write
func (fns *FileNodeSlice) Write(path string) (error) {
	// TODO
	jm, err := json.Marshal(fns)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, jm, 0644)
}

// Read Read
func (fns *FileNodeSlice) Read(path string) ([]string, error) {
	// TODO
	ret := new(FileNodeSlice)
	rf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ret.Slice, json.Unmarshal(rf, ret)
}

// walk walk
func (fns *FileNodeSlice) w(path string, wg *sync.WaitGroup, token *TokenBucket) error {
	defer token.Put()
	defer wg.Done()
	f, err := fns.readDirNames(path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	tmp := make([]string, 0, len(f))

	for _, d := range f {
		filename := filepath.Join(path, d)
		info, err := os.Lstat(filename)
		if err != nil {
			continue
		}
		if info.IsDir() {
			_, ok := fns.ig.dir[d]
			if ok {
				continue
			}
			wg.Add(1)
			token.Get()
			go func(path string) {
				fns.w(path, wg, token)
			}(filename)
		} else {
			_, ok := fns.ig.file[d]
			if ok {
				continue
			}

			tmps := strings.TrimPrefix(filename, fns.rootPath)
			tmp = append(tmp, tmps)
		}
	}
	fns.Adds(tmp...)
	return nil
}

func (fns *FileNodeSlice) readDirNames(dirname string) ([]string, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	return names, nil
}
