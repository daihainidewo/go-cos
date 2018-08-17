// Package sysfile sysfile
// file create by daihao, time is 2018/8/14 19:53
package sysfile

import (
	"encoding/json"
	"io/ioutil"
	"golang/entity"
)

// AddBuffer
type AddBuffer struct {
	path  string
	paths []string
}

// NewAddBuffer NewAddBuffer
func NewAddBuffer(path string) *AddBuffer {
	// TODO
	return &AddBuffer{
		path:  path,
		paths: make([]string, 0),
	}
}

// GetPaths GetPaths
func (ab *AddBuffer) GetPaths() ([]string) {
	// TODO
	return ab.paths
}

// SetPaths SetPaths
func (ab *AddBuffer) SetPaths(paths []string) {
	// TODO
	ab.paths = paths
}

// AddPaths AddPaths
func (ab *AddBuffer) AddPaths(paths ...string) {
	// TODO
	ab.paths = append(ab.paths, paths...)
}

// Write Write
func (ab *AddBuffer) Write() (error) {
	// TODO
	content := new(entity.AddBuffer)
	content.Paths = ab.paths
	bb, err := json.Marshal(content)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(ab.path, bb, 0644)
}

// Read Read
func (ab *AddBuffer) Read() (error) {
	// TODO
	bb, err := ioutil.ReadFile(ab.path)
	if err != nil {
		return err
	}
	ret := new(entity.AddBuffer)
	err = json.Unmarshal(bb, ret)
	if err != nil {
		return err
	}
	ab.paths = ret.Paths
	return nil
}
