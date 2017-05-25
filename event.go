package pharoscache

import "time"

const (
	event_act_set  = iota
	event_act_get
	event_act_close
)

type event struct {
	key string
	value []byte
	expire time.Duration
	action int
	resultsChan chan result
}


type result struct {
	ok bool
	value []byte
}


