// Package backup backup
// file create by daihao, time is 2018/8/10 11:59
package backup

import (
	"io/ioutil"
	"fmt"
	"encoding/json"
)

type Backup struct {
	path string
	Content   map[string]interface{}
}

// NewBackup NewBackup
func NewBackup(path string) *Backup {
	ret := &Backup{
		path: path,
		Content:   make(map[string]interface{}),
	}
	return ret
}

// Close Close
func (b *Backup) Close() {

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
