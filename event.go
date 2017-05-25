package pharoscache

const (
	event_act_set  = iota
	event_act_get
)

type event struct {
	key string
	value []byte
	action int
	resultsChan chan result
}


type result struct {
	ok bool
	value []byte
}


