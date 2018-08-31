// Package main main
// file create by daihao, time is 2018/8/30 11:46
package main

import (
	"encoding/json"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"log"
	"path/filepath"
)

type TXConf struct {
	SecretId  string `toml:"secret_id,omitempty"`
	SecretKey string `toml:"secret_key,omitempty"`
	AppId     string `toml:"app_id,omitempty"`
	Host      string `toml:"host,omitempty"`
}

// NewTXConf new TXConf
func NewTXConf(path string) *TXConf {
	ret := new(TXConf)
	_, err := toml.DecodeFile(path, ret)
	if err != nil {
		log.Fatalf("read txconf error:%s", err)
		return nil
	}
	if ret.SecretId == "" || ret.SecretKey == "" || ret.AppId == "" || ret.Host == "" {
		log.Fatalf("conf.toml file not configure message")
		return nil
	}
	return ret
}

type BackupConf struct {
	LastUpdateTime int64 `json:"last_update_time"`
}

// NewBackupConf new BackupConf
func NewBackupConf(path string) *BackupConf {
	body, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("read backup conf error:%s", err)
		return nil
	}
	ret := new(BackupConf)
	err = json.Unmarshal(body, ret)
	if err != nil {
		log.Fatalf("json unmarshal backup conf error:%s", err)
		return nil
	}
	return ret
}

// Write Write
func (b *BackupConf) Write(path string) (error) {
	jm, err := json.Marshal(b)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, jm, 0644)
}

// Sysfile
type Sysfile struct {
	projPath string // 配置文件目录
	rootpath string // 文件目录
}

// SetCosPath SetCosPath
func NewSysfile(rootpath string) (*Sysfile) {
	return &Sysfile{
		rootpath: rootpath,
		projPath: filepath.Join(rootpath, ".gocos"),
	}
}

// RootPath RootPath
func (s *Sysfile) RootPath() string {
	return s.rootpath
}

// CosPath CosPath
func (s *Sysfile) ProjPath() string {
	return s.projPath
}

// AddBufferPath AddBufferPath
func (s *Sysfile) AddBufferPath() string {
	return filepath.Join(s.projPath, ".addbuffer")
}

// BackupPath BackupPath
func (s *Sysfile) BackupPath() (string) {
	return filepath.Join(s.projPath, ".backup")
}

// ConfPath ConfPath
func (s *Sysfile) ConfPath() string {
	return filepath.Join(s.projPath, "conf.toml")
}

// IgnorePath IgnorePath
func (s *Sysfile) IgnorePath() string {
	return filepath.Join(s.projPath, "ignore")
}

// FileNodeSlice FileNodeSlice
func (s *Sysfile) FileNodeSlice() (string) {
	return filepath.Join(s.projPath, ".filenodeslice")
}

// AddBuffer
type AddBuffer struct {
	Del []string `json:"delete"`
	Crt []string `json:"crteate"`
	Mod []string `json:"modify"`
}

// NewAddBuffer NewAddBuffer
func NewAddBuffer(path string) *AddBuffer {
	ret := &AddBuffer{
		Del: make([]string, 0),
		Mod: make([]string, 0),
		Crt: make([]string, 0),
	}
	rf, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("read backup conf error:%s", err)
		return nil
	}
	err = json.Unmarshal(rf, ret)
	if err != nil {
		log.Fatalf("read backup conf error:%s", err)
		return nil
	}
	return ret
}

// AddDelPath AddDelPath
func (ab *AddBuffer) AddDelPath(path ...string) () {
	ab.Del = append(ab.Del, path...)
}

// AddModPath AddModPath
func (ab *AddBuffer) AddModPath(path ...string) {
	ab.Mod = append(ab.Mod, path...)
}

// AddPaths AddPaths
func (ab *AddBuffer) AddCrtPath(path ...string) {
	ab.Crt = append(ab.Crt, path...)
}

// Write Write
func (ab *AddBuffer) Write(path string) (error) {
	bb, err := json.Marshal(ab)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, bb, 0644)
}
