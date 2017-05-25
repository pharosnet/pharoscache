package pharoscache

import (
	"time"
	"container/list"
)

type entry struct {
	key   string
	value []byte
	expire time.Time
}

type bucket struct {
	maxEntries 	int
	evicted 	Evicted
	ll    		*list.List
	data		map[string]*list.Element
	eventChan 	chan event
}

func (b *bucket) Get(key string) ([]byte, bool) {
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

func (b *bucket) Set(key string, value []byte, expire time.Duration) {
	b.eventChan <- event{key:key, action:event_act_set, value:value, expire:expire}
}

func newBucket(maxEntries int, evicted Evicted) *bucket {
	b := &bucket{
		maxEntries: 	maxEntries,
		evicted:	evicted,
		ll:         	list.New(),
		data:      	make(map[string]*list.Element),
		eventChan: 	make(chan event, maxEntries),
	}
	b.handleEvent()
	return b
}

func (b *bucket) handleEvent() {
	go func(b *bucket) {
		for {
			evt, ok := <- b.eventChan
			if !ok {
				break
			}
			switch evt.action {
			case event_act_close:
				close(b.eventChan)
				break
			case event_act_set:
				b.set(evt.key, evt.value, evt.expire)
			case event_act_get:
				b.get(evt.key, evt.resultsChan)
			}
		}
	}(b)
}

func (b *bucket) set(key string, value []byte, expire time.Duration) {
	expireTime := time.Time{}
	if expire > 0 {
		expireTime = time.Now().Add(expire)
	}
	if ee, ok := b.data[key]; ok {
		if expire == 0 {
			b.ll.MoveToFront(ee)
		}
		ee.Value.(*entry).value = value
		ee.Value.(*entry).expire = expireTime
		return
	}
	var ele *list.Element
	if expire == 0 {
		ele = b.ll.PushFront(&entry{key:key, value:value, expire:expireTime})
		if b.maxEntries != 0 && b.ll.Len() > b.maxEntries {
			b.removeOldest()
		}
	} else {
		ele = &list.Element{Value:&entry{key:key, value:value}}
	}
	b.data[key] = ele
}

func (b *bucket) get(key string, resultsChan chan result) {
	if b.data == nil {
		resultsChan <- result{value:nil, ok:false}
		return
	}
	if ele, hit := b.data[key]; hit {
		e := ele.Value.(*entry)
		if e.expire.IsZero() {
			resultsChan <- result{value:e.value, ok:true}
			b.ll.MoveToFront(ele)
		} else {
			if e.expire.Before(time.Now()) {
				delete(b.data, key)
				if b.evicted != nil {
					b.evicted.On(key, e.value)
				}
				resultsChan <- result{value:nil, ok:false}
			} else {
				resultsChan <- result{value:e.value, ok:true}
			}
		}
	} else {
		resultsChan <- result{value:nil, ok:false}
	}
}

func (b *bucket) removeOldest() {
	if b.data == nil || len(b.data) == 0 {
		return
	}
	ele := b.ll.Back()
	if ele != nil {
		b.removeElement(ele)
	}
}

func (b *bucket) removeElement(e *list.Element) {
	b.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(b.data, kv.key)
	if b.evicted != nil {
		b.evicted.On(kv.key, kv.value)
	}
}

func (b *bucket) stop() {
	b.eventChan <- event{action:event_act_close}
}