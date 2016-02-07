package jqeventrouter

import (
	"strings"
// 	"sync"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
// 	"github.com/armon/go-radix"
	"honnef.co/go/js/console"
)

type EventMux struct {
// 	mu      sync.RWMutex
//	t       *radix.Tree
	paths   map[string]*muxEntry
	uriFunc func(*jquery.Event, *js.Object) string
}

type muxEntry struct {
	h       Handler
	pattern string
	params  string
}

func NewEventMux() *EventMux {
	return &EventMux{paths: make(map[string]*muxEntry)}
}


func (mux *EventMux) HandleEvent(event *jquery.Event, data *js.Object) bool {
	uri := mux.getUri(event, data)
	console.Log("URI = %s", uri)
	for path,entry := range mux.paths {
		if strings.HasPrefix(uri,path) {
			// We found a match!
			return entry.h.HandleEvent(event, data)
		}
	}
	return true
}

// SetUriFunc allows you to specify a custom function to determine the URI of a request.
// The custom function is expected to return a string, representing the URI.
// If unset, everything after the hostname in window.location.href is retunred.
func (mux *EventMux) SetUriFunc(fn func(event *jquery.Event, data *js.Object) string) {
	mux.uriFunc = fn
}

func (mux *EventMux) getUri(event *jquery.Event, data *js.Object) string {
	if mux.uriFunc != nil {
		return mux.uriFunc(event, data)
	}
	return js.Global.Get("location").Get("href").String()
}

// Handle registers the handler for the given pattern. If a handler already 
// exists for pattern, Handle panics.
func (mux *EventMux) Handle(pattern string, handler Handler) {
	if pattern == "" {
		panic("eventrouter: invalid pattern " + pattern)
	}
	if handler == nil {
		panic("eventrouter: nil handler")
	}
	if _,ok := mux.paths[pattern]; ok {
		panic("eventrouter: multiple registrations for " + pattern)
	}
	
	var params string

	// Check for any named parameters
	if i := strings.LastIndexAny(pattern,":*"); i > -1 {
		if j := strings.LastIndex(pattern, "/"); j < i {
			params = pattern[i:]
			pattern = pattern[0:i-1]
		}
	}

	mux.paths[pattern] = &muxEntry{
		h: handler,
		pattern: pattern,
		params: params,
	}
}

func (mux *EventMux) HandleFunc(pattern string, handler func(*jquery.Event, *js.Object) bool) {
	mux.Handle(pattern, HandlerFunc(handler))
}

// Listen is a convenience function which calls Listen(event,mux)
func (mux *EventMux) Listen(event string) {
	Listen(event, mux)
}

type Handler interface {
	HandleEvent(event *jquery.Event, data *js.Object) bool
}

// The HandlerFunc type is an adaptor to allow the use of ordinary functions
// as Event handlers. If f is a function of the appropriate signature, HandlerFunc(f)
// is a Handler object that calls f.
type HandlerFunc func(*jquery.Event, *js.Object) bool

// HandleEvent calls f(this,event)
func (f HandlerFunc) HandleEvent(event *jquery.Event, data *js.Object) bool {
	return f(event, data)
}

type EventListener struct {
	event    string
	listener func(*jquery.Event, *js.Object) bool
	detached bool
}

// Listen attaches the Handler to the window and begins listening for the specified
// jquery event, reterning an EventListener object
func Listen(event string, handler Handler) *EventListener {
	console.Log("Adding jQuery event listener")
	listener := func(event *jquery.Event, data *js.Object) bool {
		console.Log("listener")
		return handler.HandleEvent(event, data)
	}
	jquery.NewJQuery(js.Global.Get("document")).On(event, listener)
	return &EventListener{event: event, listener: listener}
}

// UnListen detaches the EventListener from the window. 
func (l *EventListener) UnListen() {
	if l.detached == true {
		panic("Already detached")
	}
	jquery.NewJQuery(js.Global.Get("document")).Off(l.event, l.listener)
	l.detached = true
}

// NullHandler returns a null Handler object, which unconditionally returns true.
// It can be used for terminating event chains that don't need to affect the page.
func NullHandler() Handler {
	return HandlerFunc(func(_ *jquery.Event, _ *js.Object) bool {
		return true
	})
}
