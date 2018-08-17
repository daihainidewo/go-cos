// Package sysfile sysfile
// file create by daihao, time is 2018/8/10 11:59
package sysfile

import (
	"io/ioutil"
	"fmt"
	"encoding/json"
)

type Backup struct {
	path    string
	Content *map[string]string
}

// NewBackup NewBackup
func NewBackup(path string) *Backup {
	ret := &Backup{
		path: path,
	}
	m := make(map[string]string)
	ret.Content = &m
	return ret
}

// Close Close
func (b *Backup) Close() {

}

// Get Get
func (b *Backup) Get(name string) string {
	return (*b.Content)[name]
}

// Set Set
func (b *Backup) Set(name, data string) {
	(*b.Content)[name] = data
}

// Read Read
func (b *Backup) Read() error {
	bb, err := ioutil.ReadFile(b.path)
	if err != nil {
		fmt.Print(err)
	}
	return json.Unmarshal(bb, b.Content)
}

// Write Write
func (b *Backup) Write() error {
	bb, err := json.Marshal(b.Content)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(b.path, bb, 0644)
}
