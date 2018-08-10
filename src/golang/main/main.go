// Package main main
// file create by daihao, time is 2018/8/10 10:28
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"golang/tool"
	"path/filepath"
	"golang/backup"
	"golang/entity"
	"encoding/json"
)

// main main
func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please enter the command parameters, The system only supports add|checkout|init|pull|push|status operation.")
		return
	}
	if os.Args[1] == "init" {
		initsys()
		return
	}
	sysdir := "/.gocos"
	tp, err := filepath.Abs(".")
	if err != nil {
		fmt.Println(err)
		return
	}
	tpp := ""
	projpath := ""
	for {
		if tpp == tp {
			break
		}
		if tool.IsDir(tp + sysdir) {
			projpath = tp + sysdir
			break
		} else {
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

	backup.Sysdata = backup.NewSysData(projpath)
	err = backup.Sysdata.Ig.Read(projpath + "/ignore")
	if err != nil {
		fmt.Println("read ignore err", err)
		return
	}

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
	default:
		fmt.Println("The system only supports add|checkout|init|pull|push|status operation.")
		return
	}
	signCh <- syscall.SIGTERM
}

// init init
func initsys() {
	if tool.IsDir(".gocos") {
		if len(os.Args) > 2 && os.Args[2] == "-f" {

		} else {
			fmt.Println("go-cos initialized, if you still want to initialize, please use 'init -f', or delete .gocos/.")
			return
		}
	} else {
		err := os.Mkdir(".gocos", 0711)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	f, err := os.Create(".gocos/conf.yaml")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	f.Write([]byte("# system init"))
	f, err = os.Create(".gocos/ignore")
	if err != nil {
		fmt.Println(err)
		return
	}
	f.Write([]byte(".gocos/"))
	f, err = os.Create(".gocos/backup.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	f.Write([]byte("{}"))
}

// status status
func status() {
	bu := backup.NewBackup(backup.Sysdata.GetProjectPath() + "/backup.json")
	bu.Read()
	ts, ok := bu.Content[entity.LastUpdateTimestamp]
	if !ok {
		ts = int64(0)
	}

	res, err := tool.CheckDirUpdate(filepath.Dir(backup.Sysdata.GetProjectPath()), ts.(int64))
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
	if len(os.Args) < 3 {
		fmt.Println("Please enter the path you want to add.")
	}
	bu := backup.NewBackup(backup.Sysdata.GetProjectPath() + "/backup.json")
	ts, ok := bu.Content[entity.LastUpdateTimestamp]
	if !ok {
		ts = int64(0)
	}
	tsn := ts.(int64)
	paths := make([]string, 0)
	for _, d := range os.Args[2:] {
		d, err := filepath.Abs(d)
		if err != nil {
			fmt.Println(err)
			continue
		}
		res, err := tool.CheckDirUpdate(d, tsn)
		if err != nil {
			fmt.Println(err)
			return
		}
		paths = append(paths, res...)
	}

	ab := &entity.AddBuffer{Paths: paths}
	ret, err := json.Marshal(ab)
	if err != nil {
		fmt.Println(err)
		return
	}
	f, err := os.Create(backup.Sysdata.GetProjectPath() + "/add-buffer")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	n, err := f.Write(ret)
	if err != nil {
		fmt.Println(n, err)
	}
}

// push push
func push() {
	// TODO

}

// pull pull
func pull() {
	// TODO

}

// checkout checkout
func checkout() {
	// TODO

}
