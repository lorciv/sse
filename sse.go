// Package sse implements a http.Handler capable of delivering server-side events to clients.
package sse

import (
	"fmt"
	"log"
	"net/http"
)

// Stream is an HTTP event stream.
// It implements the http.Handler interface, through which clients can subscribe to new events.
// It exposes Send methods, that the server can use to notify all active listeners.
type Stream struct {
	requests chan request
	channels []chan message

	Logger *log.Logger
}

// NewStream returns a new event stream that is ready to use.
func NewStream() *Stream {
	s := Stream{requests: make(chan request)}
	go s.run()
	return &s
}

// Send sends a new event to all listening clients.
// It is equivalent to a call to SendEvent with event == "message".
func (s *Stream) Send(data string) {
	s.SendEvent("message", data)
}

// SendEvent sends a new event to all listening clients, specifying the event type.
func (s *Stream) SendEvent(event, data string) {
	s.requests <- request{cmd: "notify", m: message{event: event, data: data}}
}

func (s *Stream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, ok := w.(http.Flusher); !ok {
		http.Error(w, "server-side events not supported", http.StatusNotImplemented)
		return
	}

	messages := s.subscribe()
	defer func() {
		s.leave(messages)
	}()

	w.Header().Set("Content-Type", "text/event-stream")

	for m := range messages {
		var err error

		_, err = fmt.Fprintf(w, "event: %s\n", m.event)
		if err != nil {
			s.logf("connection lost")
			break
		}

		_, err = fmt.Fprintf(w, "data: %s\n\n", m.data)
		if err != nil {
			s.logf("connection lost")
			break
		}

		w.(http.Flusher).Flush()
	}
}

type request struct {
	cmd string       // one of "subscribe", "leave", "notify"
	c   chan message // only for "subscribe", "leave"
	m   message      // only for "notify"
}

type message struct {
	event, data string
}

func (s *Stream) run() {
	for req := range s.requests {
		switch req.cmd {
		case "subscribe":
			s.channels = append(s.channels, req.c)
			s.logf("new subscriber: total %d", len(s.channels))
		case "leave":
			for i, c := range s.channels {
				if c == req.c {
					close(c)

					// Remove from list
					last := len(s.channels) - 1
					s.channels[i] = s.channels[last]
					s.channels = s.channels[:last]

					break
				}
			}
			s.logf("del subscriber: total %d", len(s.channels))
		case "notify":
			for _, c := range s.channels {
				c <- req.m
			}
		default:
			panic("unexpected request type")
		}
	}
}

func (s *Stream) subscribe() chan message {
	c := make(chan message)
	s.requests <- request{cmd: "subscribe", c: c}
	return c
}

func (s *Stream) leave(c chan message) {
	s.requests <- request{cmd: "leave", c: c}
	for range c {
		// Drain the channel
	}
}

func (s *Stream) logf(format string, v ...any) {
	if s.Logger != nil {
		s.Logger.Printf(format, v...)
	}
}
