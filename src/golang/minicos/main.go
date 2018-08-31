// Package minicos minicos
// file create by daihao, time is 2018/8/30 11:43
package main

import (
	"bytes"
	"fmt"
	"github.com/labstack/gommon/log"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

const (
	LastUpdateTimestamp = "LastUpdateTimestamp"
	TokenBucketSize     = 50
)

var syscos *Sysfile

func init() {
}

func main() {
	start := time.Now()
	defer func() {
		end := time.Now()
		fmt.Printf("Operating time %gs\n", end.Sub(start).Seconds())
	}()
	if len(os.Args) < 2 {
		fmt.Println("Please enter the command parameters, The system only supports add|checkout|init|pull|push|status operation.")
		return
	}

	tp, err := filepath.Abs(".")
	if err != nil {
		fmt.Println(err)
		return
	}

	if os.Args[1] == "init" {
		syscos = NewSysfile(tp)
		initsys()
		return
	}

	tpp := filepath.VolumeName(tp)
	projpath := ""
	for {
		projpath = filepath.Join(tp, ".gocos")
		info, _ := os.Lstat(projpath)
		if info.IsDir() {
			break
		} else {
			if tpp == tp {
				projpath = ""
				break
			}
			tp = filepath.Dir(tp)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

	}
	if projpath == "" {
		fmt.Println("not found .gocos")
		return
	}

	syscos = NewSysfile(tp)

	signCh := make(chan os.Signal)
	signal.Notify(signCh, os.Interrupt, os.Kill, syscall.SIGTERM)
	go startup(os.Args, signCh)
	<-signCh
}

// startup startup
func startup(args []string, signCh chan os.Signal) {
	switch args[1] {
	case "status":
		status()
	case "add":
		add()
	case "push":
		push()
	case "pull":
		pull()
	case "checkout":
		checkout()
	case "clear":
		clear()
	default:
		fmt.Println("The system only supports add|checkout|init|pull|push|status operation.")
	}
	signCh <- syscall.SIGTERM
}

// init init
func initsys() {
	info, _ := os.Lstat(syscos.ProjPath())
	if info.IsDir() {
		if len(os.Args) > 2 && os.Args[2] == "-f" {

		} else {
			fmt.Println("go-cos initialized, if you still want to initialize, please use 'init -f', or delete .gocos/.")
			return
		}
	} else {
		err := os.Mkdir(syscos.ProjPath(), 0711)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	f, err := os.Create(syscos.ConfPath())
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	f.Write([]byte("# system init"))

	f, err = os.Create(syscos.IgnorePath())
	if err != nil {
		fmt.Println(err)
		return
	}
	f.Write([]byte(".gocos/"))

	nab := &AddBuffer{
		Del: make([]string, 0),
		Mod: make([]string, 0),
		Crt: make([]string, 0),
	}
	err = nab.Write(syscos.AddBufferPath())
	if err != nil {
		log.Errorf("write add buffer file error:%s", )
	}

	bk := &BackupConf{
		LastUpdateTime: 0,
	}
	err = bk.Write(syscos.BackupPath())
	if err != nil {
		log.Errorf("write backup file error:%s", err)
	}

	fn := &FileNodeSlice{
		Slice: []string{},
	}
	err = fn.Write(syscos.FileNodeSlice())
	if err != nil {
		log.Errorf("write file node slice file error:%s", err)
	}
}

// status status
func status() {
	fns := NewFileNodeSlice(syscos.RootPath(), NewIgnore(syscos.IgnorePath()))
	last, err := fns.Read(syscos.FileNodeSlice())
	if err != nil {
		log.Errorf("read file node slice error:%s", err)
		return
	}
	bc := NewBackupConf(syscos.BackupPath())

	fns.Walk(syscos.RootPath())

	c, m, d := fns.Check(last, bc.LastUpdateTime)
	for _, t := range c {
		fmt.Println("create", t)
	}
	for _, t := range m {
		fmt.Println("modify", t)
	}
	for _, t := range d {
		fmt.Println("delete", t)
	}

	fmt.Println("file num:", len(c)+len(m)+len(d))
}

// add add
func add() {
	fns := NewFileNodeSlice(syscos.RootPath(), NewIgnore(syscos.IgnorePath()))

	last, err := fns.Read(syscos.FileNodeSlice())
	if err != nil {
		log.Errorf("read file node slice error:%s", err)
		return
	}

	bc := NewBackupConf(syscos.BackupPath())

	fns.Walk(syscos.RootPath())

	c, m, d := fns.Check(last, bc.LastUpdateTime)
	ab := AddBuffer{
		Crt: c,
		Mod: m,
		Del: d,
	}
	ab.Write(syscos.AddBufferPath())
}

// push push
func push() {

	ab := NewAddBuffer(syscos.AddBufferPath())
	tc := NewTXConf(syscos.ConfPath())
	txc := NewTXcos(tc)
	ig := NewIgnore(syscos.IgnorePath())
	fn := NewFileNodeSlice(syscos.RootPath(), ig)

	tb := NewTokenBucket(TokenBucketSize)
	wg := new(sync.WaitGroup)
	sm := new(sync.Mutex)
	success := 0
	loss := 0
	//errpath := make([]string, 0)
	tb.Get()
	wg.Add(1)
	go func() {
		defer func() {
			tb.Put()
			wg.Done()
		}()
		for _, path := range ab.Crt {
			tb.Get()
			wg.Add(1)
			go func(path string) {
				defer func() {
					tb.Put()
					wg.Done()
				}()

				abspath := filepath.Join(syscos.RootPath(), path)
				data, err := ioutil.ReadFile(abspath)
				if err != nil {
					log.Errorf("open file error, path", abspath)
					return
				}

				body := bytes.NewBuffer(data)
				resp, err := txc.NewRequest("PUT", path, url.Values{}, http.Header{}, body).SetHeaderContentLength(strconv.Itoa(body.Len())).Do()
				if err != nil {
					sm.Lock()
					loss++
					//errpath = append(errpath, path)
					sm.Unlock()
					log.Errorf("txcos put %s, error:%s", path, err)
					return
				}
				fmt.Println("put new file " + path + " --> " + resp.Status)

				sm.Lock()
				success++
				sm.Unlock()
			}(path)
		}
	}()

	tb.Get()
	wg.Add(1)
	go func() {
		defer func() {
			tb.Put()
			wg.Done()
		}()
		for _, path := range ab.Mod {
			tb.Get()
			wg.Add(1)
			go func(path string) {
				defer func() {
					tb.Put()
					wg.Done()
				}()

				abspath := filepath.Join(syscos.RootPath(), path)
				data, err := ioutil.ReadFile(abspath)
				if err != nil {
					log.Errorf("open file error, path", abspath)
					return
				}

				body := bytes.NewBuffer(data)
				resp, err := txc.NewRequest("PUT", path, url.Values{}, http.Header{}, body).SetHeaderContentLength(strconv.Itoa(body.Len())).Do()
				if err != nil {
					sm.Lock()
					loss++
					//errpath = append(errpath, path)
					sm.Unlock()
					log.Errorf("txcos put %s, error:%s", path, err)
					return
				}
				fmt.Println("put modify file " + path + " --> " + resp.Status)

				sm.Lock()
				success++
				sm.Unlock()
			}(path)
		}
	}()

	tb.Get()
	wg.Add(1)
	go func() {
		defer func() {
			tb.Put()
			wg.Done()
		}()
		for _, path := range ab.Del {
			tb.Get()
			wg.Add(1)
			go func(path string) {
				defer func() {
					tb.Put()
					wg.Done()
				}()

				resp, err := txc.NewRequest("DELETE", path, url.Values{}, http.Header{}, nil).Do()
				if err != nil {
					sm.Lock()
					loss++
					//errpath = append(errpath, path)
					sm.Unlock()
					log.Errorf("txcos put %s, error:%s", path, err)
					return
				}
				fmt.Println("delete file " + path + " --> " + resp.Status)

				sm.Lock()
				success++
				sm.Unlock()
			}(path)
		}
	}()

	nab := &AddBuffer{
		Del: make([]string, 0),
		Mod: make([]string, 0),
		Crt: make([]string, 0),
	}
	err := nab.Write(syscos.AddBufferPath())
	if err != nil {
		log.Errorf("write add buffer file error:%s", )
	}

	fn.Walk(syscos.RootPath())

	err = fn.Write(syscos.FileNodeSlice())
	if err != nil {
		log.Errorf("write file node slice file error:%s", err)
	}

	wg.Wait()
	if loss == 0 {
		tn := time.Now().Unix()
		bk := &BackupConf{
			LastUpdateTime: tn,
		}
		err = bk.Write(syscos.BackupPath())
		if err != nil {
			log.Errorf("write backup file error:%s", err)
		}
	}

	fmt.Println()
	fmt.Println("success:", success, ", loss:", loss)
}

// pull pull
func pull() {
	// TODO
	fmt.Println("Temporarily not supported")
}

// checkout checkout
func checkout() {
	// TODO
	fmt.Println("Temporarily not supported")
}

// clear clear
func clear() {
	os.RemoveAll(syscos.ProjPath())
	fmt.Println("clear .gocos over")
}
