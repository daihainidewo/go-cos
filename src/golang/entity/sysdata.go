// Package entity entity
// file create by daihao, time is 2018/8/10 13:35
package entity

const (
	LastUpdateTimestamp = "LastUpdateTimestamp"
)

// AddBuffer
type AddBuffer struct {
	Paths []string `json:"paths"`
}
