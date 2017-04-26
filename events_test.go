package slack

import (
	"reflect"
	"testing"
)

type TestHandler struct{}
func (t *TestHandler) ServeEvent(msg *Message, slack *Client){}

func TestHandle(t *testing.T) {
	mux := NewEventMux()
	handler := &TestHandler{}
	mux.Handle("message", handler)

	actual := mux.m["message"]
	expected := muxEntry{explicit: true, handler: handler, event: "message"}

	if actual != expected {
		t.Errorf("Failed")
	}
}

func TestEventMatch(t *testing.T) {
	actual := eventMatch("message", "message")
	if actual != true {
		t.Errorf("Failed")
	}
}

func TestMatch(t *testing.T) {
	mux := NewEventMux()
	handler := &TestHandler{}
	mux.m = make(map[string]muxEntry)
	mux.m["message"] = muxEntry{explicit: true, handler: handler, event: "message"}
	actual, _ := mux.match("message")
	match := reflect.DeepEqual(actual, handler)
	if !match {
		t.Errorf("Failed")
	}
}
