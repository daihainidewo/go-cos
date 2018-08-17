// Package tool tool
// file create by daihao, time is 2018/8/10 10:51
package tool

type TokenBucket chan struct{}

func NewTokenBucket(size int) TokenBucket {
	tb := make(chan struct{}, size)
	for i := 0; i < size; i++ {
		tb <- struct{}{}
	}

	return TokenBucket(tb)
}

func (tb TokenBucket) Get() {
	<-tb
}

func (tb TokenBucket) Put() {
	tb <- struct{}{}
}
