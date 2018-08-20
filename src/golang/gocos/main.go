// Package gocos gocos
// file create by daihao, time is 2018/8/10 10:28
package gocos

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"golang/tool"
	"path/filepath"
	"golang/entity"
	"encoding/json"
	"io/ioutil"
	"bytes"
	"golang/sysfile"
	"strconv"
	"net/http"
	"golang/cos"
	"github.com/BurntSushi/toml"
	"sync"
	"strings"
	"time"
	"golang/conf"
)

const (
	LastUpdateTimestamp = "LastUpdateTimestamp"
	TokenBucketSize     = 20
)

var (
	SysCos *sysfile.Sysfile
)

// gocos gocos
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
		SysCos = sysfile.NewSysfile(tp)
		initsys()
		return
	}

	tpp := ""
	projpath := ""
	for {
		projpath = tool.PathLink(tp, ".gocos")
		if tool.IsDir(projpath) {
			break
		} else {
			if tpp == tp {
				break
			}
			tpp = tp
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

	SysCos = sysfile.NewSysfile(tp)

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
	if tool.IsDir(SysCos.ProjPath()) {
		if len(os.Args) > 2 && os.Args[2] == "-f" {

		} else {
			fmt.Println("go-cos initialized, if you still want to initialize, please use 'init -f', or delete .gocos/.")
			return
		}
	} else {
		err := os.Mkdir(SysCos.ProjPath(), 0711)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	f, err := os.Create(SysCos.ConfPath())
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	f.Write([]byte("# system init"))
	f, err = os.Create(SysCos.IgnorePath())
	if err != nil {
		fmt.Println(err)
		return
	}
	f.Write([]byte(".gocos/"))
	f, err = os.Create(SysCos.SysfilePath())
	if err != nil {
		fmt.Println(err)
		return
	}

	bu := sysfile.NewBackup(SysCos.SysfilePath())
	bu.Set(LastUpdateTimestamp, "0")
	err = bu.Write()
	if err != nil {
		fmt.Println("write sysfile error", err)
		return
	}
}

// status status
func status() {
	ig := sysfile.NewIgnore(SysCos.IgnorePath())
	err := ig.Read()
	if err != nil {
		fmt.Println("read ignore error", err)
		return
	}

	bu := sysfile.NewBackup(SysCos.SysfilePath())
	err = bu.Read()
	if err != nil {
		fmt.Println("read sysfile error", err)
		return
	}
	lut, err := strconv.ParseInt(bu.Get(LastUpdateTimestamp), 10, 64)
	if err != nil {
		fmt.Println("parse sysfile error", err)
	}

	res, err := tool.CheckDirUpdate(ig, SysCos.RootPath(), lut)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Changed:", len(res))
	for _, d := range res {
		fmt.Println(d)
	}
}

// add add
func add() {
	ig := sysfile.NewIgnore(SysCos.IgnorePath())
	if ig.Read() != nil {
		fmt.Println("read ignore error")
		return
	}

	bu := sysfile.NewBackup(SysCos.SysfilePath())
	if bu.Read() != nil {
		fmt.Println("read sysfile error")
		return
	}
	ts, err := strconv.ParseInt(bu.Get(LastUpdateTimestamp), 10, 64)
	if err != nil {
		fmt.Println("parse sysfile error", err)
	}
	paths := make([]string, 0)

	res, err := tool.CheckDirUpdate(ig, SysCos.RootPath(), ts)
	if err != nil {
		fmt.Println(err)
		return
	}
	for i, d := range res {
		res[i] = strings.TrimPrefix(d, SysCos.RootPath()+"/")
	}
	paths = append(paths, res...)

	ab := &entity.AddBuffer{Paths: paths}
	ret, err := json.Marshal(ab)
	if err != nil {
		fmt.Println(err)
		return
	}
	f, err := os.Create(SysCos.AddBufferPath())
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	n, err := f.Write(ret)
	if err != nil {
		fmt.Println(n, err)
	}
	fmt.Println("add file", len(paths))
}

// push push
func push() {
	pathss, err := ioutil.ReadFile(SysCos.AddBufferPath())
	if err != nil {
		fmt.Println(err)
		return
	}
	paths := new(entity.AddBuffer)
	err = json.Unmarshal(pathss, paths)
	if err != nil {
		fmt.Println("please exec add again,", err)
		return
	}
	cf := new(conf.Conf)
	_, err = toml.DecodeFile(SysCos.ConfPath(), cf)
	if err != nil {
		fmt.Println("toml decode conf.toml error")
		return
	}
	tx := cos.NewTXcos(cf.SecretId, cf.SecretKey, cf.AppId, cf.Host)
	tb := tool.NewTokenBucket(TokenBucketSize)
	wg := new(sync.WaitGroup)
	sm := new(sync.Mutex)
	success := 0
	loss := 0
	errpath := make([]string, 0)
	for _, path := range paths.Paths {
		tb.Get()
		wg.Add(1)
		go func(path string) {
			defer func() {
				tb.Put()
				wg.Done()
			}()
			abspath := tool.PathLink(SysCos.RootPath(), path)
			data, err := ioutil.ReadFile(abspath)
			if err != nil {
				fmt.Println("open file error, path", abspath)
				return
			}
			body := bytes.NewBuffer(data)
			headers := http.Header{}
			headers.Set("Content-Length", strconv.Itoa(body.Len()))
			resp, err := tx.Sendfile("/"+path, headers, body)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(path + " --> " + resp.Status)
			if resp.StatusCode >= 400 {
				defer resp.Body.Close()
				html, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(string(html))
				sm.Lock()
				loss++
				errpath = append(errpath, path)
				sm.Unlock()
				return
			}
			sm.Lock()
			success++
			sm.Unlock()
		}(path)
	}

	bu := sysfile.NewBackup(SysCos.SysfilePath())
	bu.Set(LastUpdateTimestamp, strconv.Itoa(int(time.Now().Unix())))
	err = bu.Write()
	if err != nil {
		fmt.Println("write sysfile error", err)
		return
	}

	addbuf := sysfile.NewAddBuffer(SysCos.AddBufferPath())
	addbuf.SetPaths(errpath)
	err = addbuf.Write()
	if err != nil {
		fmt.Println(err)
		return
	}

	wg.Wait()
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
	os.RemoveAll(SysCos.ProjPath())
	fmt.Println("clear .gocos over")
}
