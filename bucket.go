package pharoscache

import (
	"container/list"
)

type entry struct {
	key   string
	value []byte
}

type Bucket struct {
	maxEntries 	int
	evicted 	Evicted
	ll    		*list.List
	data		map[string]*list.Element
	eventChan 	chan event
}

func (b *Bucket) Get(key string) ([]byte, bool) {
	resultsChan := make(chan result, 1)
	e := event{key:key, action:event_act_get, resultsChan:resultsChan}
	b.eventChan <- e
	rs, closed := <- resultsChan
	if closed {
		close(resultsChan)
		return rs.value, rs.ok
	}
	return nil, false
}

func (b *Bucket) Set(key string, value []byte) {
	b.eventChan <- event{key:key, action:event_act_set, value:value}
}

func newBucket(maxEntries int, evicted Evicted) *Bucket {
	b := &Bucket{
		maxEntries: 	maxEntries,
		evicted:	evicted,
		ll:         	list.New(),
		data:      	make(map[string]*list.Element),
		eventChan: 	make(chan event, maxEntries),
	}
	b.handleEvent()
	return b
}

func (b *Bucket) handleEvent() {
	go func(b *Bucket) {
		for {
			evt, ok := <- b.eventChan
			if !ok {
				break
			}
			switch evt.action {
			case event_act_set:
				b.set(evt.key, evt.value)
			case event_act_get:
				b.get(evt.key, evt.resultsChan)
			}
		}
	}(b)
}

func (b *Bucket) set(key string, value []byte) {
	if ee, ok := b.data[key]; ok {
		b.ll.MoveToFront(ee)
		ee.Value.(*entry).value = value
		return
	}
	ele := b.ll.PushFront(&entry{key, value})
	b.data[key] = ele
	if b.maxEntries != 0 && b.ll.Len() > b.maxEntries {
		b.removeOldest()
	}
}

func (b *Bucket) get(key string, resultsChan chan result) {
	if b.data == nil {
		resultsChan <- result{value:nil, ok:false}
		return
	}
	if ele, hit := b.data[key]; hit {
		resultsChan <- result{value:ele.Value.(*entry).value, ok:true}
		b.ll.MoveToFront(ele)
	} else {
		resultsChan <- result{value:nil, ok:false}
	}
}

func (b *Bucket) removeOldest() {
	if b.data == nil || len(b.data) == 0 {
		return
	}
	ele := b.ll.Back()
	if ele != nil {
		b.removeElement(ele)
	}
}

func (b *Bucket) removeElement(e *list.Element) {
	b.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(b.data, kv.key)
	if b.evicted != nil {
		b.evicted.On(kv.key, kv.value)
	}
}