package notifier

import (
	"fmt"
	"log"
	"sync"
)

type INotiManager interface {
	AddSubscriber(s ISubscriber)
	Publish(e IEvent) bool
	RemoveSubscriber(s ISubscriber)
	GetSubscriberList() map[string][]ISubscriber
}

type NotiManager struct {
	subscribers map[string][]ISubscriber
	mutex       sync.Mutex
}

func NewNotiManager() *NotiManager {
	return &NotiManager{map[string][]ISubscriber{}, sync.Mutex{}}
}

func (nm *NotiManager) GetSubscriberList() map[string][]ISubscriber {
	return nm.subscribers
}

func (nm *NotiManager) AddSubscriber(s ISubscriber) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	nm.subscribers[s.Token()] = append(nm.subscribers[s.Token()], s)
	fmt.Println(nm.subscribers)
}

func (nm *NotiManager) RemoveSubscriber(s ISubscriber) {
	sublist, ok := nm.subscribers[s.Token()]
	if !ok {
		return
	}

	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	for i, e := range sublist {
		if e == s {
			sublist[i] = sublist[len(sublist)-1]
			if len(sublist)-1 == 0 {
				delete(nm.subscribers, s.Token())
			} else {
				nm.subscribers[s.Token()] = sublist[:len(sublist)-1]
			}
		}
	}

	log.Println(nm.subscribers)
}

func (nm *NotiManager) Publish(e IEvent) bool {
	sublist, ok := nm.subscribers[e.Token()]
	if !ok {
		return ok
	}

	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	tail := len(sublist) - 1
	idx := 0
	for idx <= tail {
		sublist[idx].Handle(e)
		if sublist[idx].Type() == SubtypeOnce {
			sublist[idx], sublist[tail] = sublist[tail], sublist[idx]
			tail--
		} else {
			idx++
		}
	}

	sublist = sublist[:idx]
	if len(sublist) == 0 {
		delete(nm.subscribers, e.Token())
	}
	fmt.Println(nm.subscribers)

	return true
}
