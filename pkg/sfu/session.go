package sfu

import (
	"fmt"
	"miniSFU/log"
	"sync"
)

// Session represents a set of transports. Transports inside a session
// are automatically subscribed to each other.
type Session struct {
	id             string
	transports     map[string]Transport
	transportsLock sync.RWMutex
	onCloseHandler func()
}

// NewSession creates a new session
func NewSession(id string) *Session {
	return &Session{
		id:         id,
		transports: make(map[string]Transport),
	}
}

// AddTransport adds a transport to the session
func (r *Session) AddTransport(transport Transport) {
	r.transportsLock.Lock()
	defer r.transportsLock.Unlock()

	r.transports[transport.ID()] = transport
}

// RemoveTransport removes a transport for the session
func (r *Session) RemoveTransport(tid string) {
	r.transportsLock.Lock()
	defer r.transportsLock.Unlock()
	delete(r.transports, tid)

	// Remove transport subs from pubs
	for _, t := range r.transports {
		for _, router := range t.Routers() {
			router.DelSub(tid)
		}
	}

	// Close session if no transports
	if len(r.transports) == 0 && r.onCloseHandler != nil {
		r.onCloseHandler()
	}
}

// one track, one router
func (r *Session) AddRouter(router *Router) {
	r.transportsLock.Lock()
	defer r.transportsLock.Unlock()

	for tid, t := range r.transports {
		// Don't sub to self
		if router.tid == tid {
			continue
		}
		log.Infof("AddRouter ssrc %d to %s", router.Track().SSRC(), tid)

		// 其他transport为该track注册一个发送器
		sender, err := t.NewSender(router.Track())
		if err != nil {
			log.Errorf("Error subscribing transport to router: %s", err)
			continue
		}

		// Attach sender to source1
		router.AddSender(tid, sender)
	}
}

// Transports returns transports in this session
func (r *Session) Transports() map[string]Transport {
	r.transportsLock.RLock()
	defer r.transportsLock.RUnlock()
	return r.transports
}

// OnClose called when session is closed
func (r *Session) OnClose(f func()) {
	r.onCloseHandler = f
}

func (r *Session) stats() string {
	info := fmt.Sprintf("\nsession: %s\n", r.id)

	r.transportsLock.RLock()
	for _, transport := range r.transports {
		info += transport.stats()
	}
	r.transportsLock.RUnlock()

	return info
}