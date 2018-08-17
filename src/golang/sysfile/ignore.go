// Package sysfile sysfile
// file create by daihao, time is 2018/8/10 15:44
package sysfile

import (
	"io/ioutil"
	"bytes"
)

// Ignore
type Ignore struct {
	path   string
	igdir  []string
	igfile []string
}

// NewIgnore new Ignore
func NewIgnore(path string) *Ignore {
	return &Ignore{
		path:   path,
		igfile: make([]string, 0),
		igdir:  make([]string, 0),
	}
}

// Read Read
func (ig *Ignore) Read() error {
	b, err := ioutil.ReadFile(ig.path)
	if err != nil {
		return err
	}
	res := bytes.Split(b, []byte("\n"))
	for _, d := range res {
		d = bytes.TrimSpace(d)
		if len(d) == 0 {
			continue
		}
		if d[len(d)-1] == '/' {
			ig.AddIgdir(string(d))
		} else {
			ig.AddIgfile(string(d))
		}
	}
	return nil
}

// AddIgdir AddIgdir
func (ig *Ignore) AddIgdir(dirname string) {
	ig.igdir = append(ig.igdir, dirname)
}

// AddIgfile AddIgfile
func (ig *Ignore) AddIgfile(filename string) {
	ig.igfile = append(ig.igfile, filename)
}

// GetIgdir GetIgdir
func (ig *Ignore) GetIgdir() []string {
	return ig.igdir
}

// GetIgfile GetIgfile
func (ig *Ignore) GetIgfile() []string {
	return ig.igfile
}
