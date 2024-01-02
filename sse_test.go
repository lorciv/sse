// Package sse implements a http.Handler capable of delivering server-side events to clients.
package sse_test

import (
	"bufio"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lorciv/sse"
)

type counter struct {
	count int
	mu    sync.RWMutex
}

func (c *counter) Incr(i int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count += i
}

func (c *counter) Set(i int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.count = i
}

func (c *counter) Get() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.count
}

func TestStream_ServeHTTP(t *testing.T) {
	stream := sse.NewStream()

	// start http server
	svr := httptest.NewServer(stream)
	defer svr.Close()

	c := &counter{}

	testEvents := []Out{
		{"tc1", []byte("tc1")},
		{"tc2", []byte("tc2")},
		{"tc3", []byte("tc3")},
	}

	sort.Slice(testEvents, func(i, j int) bool {
		return testEvents[i].Event < testEvents[j].Event &&
			string(testEvents[i].Data) < string(testEvents[j].Data)
	})

	// create clients to receive the event
	var wg sync.WaitGroup
	noOfClients := 5
	for i := 0; i < noOfClients; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			client := NewClient(id, svr.URL, c)
			err := client.RecvEvents()
			if err != nil {
				log.Println(err)
			}
			client.VerfiyEventsReceived(t, testEvents)
		}(i)
	}

	time.Sleep(1 * time.Second) // allow clients to connect

	// create events in server
	for ind, tc := range testEvents {
		stream.SendEvent(tc.Event, tc.Data)

		// wait for all clients to recv
		for j := 0; j < 100; j++ {
			if c.Get() == noOfClients*(ind+1) {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}

		if c.Get() != noOfClients*(ind+1) {
			t.Errorf("Expected %d events, got %d", noOfClients*(ind+1), c.Get())
		}
	}

	stream.LeaveAll() // close all clients

	wg.Wait() // allow clients to close and verify
}

func (c Client) VerfiyEventsReceived(t *testing.T, expectedEvents []Out) {
	// Sort the array
	sort.Slice(c.recvdEvents, func(i, j int) bool {
		return c.recvdEvents[i].Event < c.recvdEvents[j].Event &&
			string(c.recvdEvents[i].Data) < string(c.recvdEvents[j].Data)
	})

	// Check if the arrays are equal
	if !reflect.DeepEqual(expectedEvents, c.recvdEvents) {
		t.Errorf("client id : %d, Expected events : %v, Received events: %v", c.id, expectedEvents, c.recvdEvents)
	}
}

type Out struct {
	Event string
	Data  []byte
}

type Client struct {
	id          int
	url         string
	recvdEvents []Out
	c           *counter
}

func NewClient(id int, url string, c *counter) *Client {
	return &Client{id: id, url: url, c: c}
}

func (c *Client) RecvEvents() error {
	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, c.url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	return c.ReadData(resp)
}

func (c *Client) ReadData(resp *http.Response) error {
	reader := bufio.NewReader(resp.Body)
	var event Out
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				if event.Event != "" || event.Data != nil {
					c.recvdEvents = append(c.recvdEvents, event)
					c.c.Incr(1)
				}
				return nil
			}
			return err
		}

		sl := string(line)
		// populate this into c.recvdEvents struct
		if strings.HasPrefix(sl, "event: ") {
			if event.Event != "" {
				c.recvdEvents = append(c.recvdEvents, event)
				c.c.Incr(1)
				event = Out{}
			}
			event.Event = sl[len("event: ") : len(sl)-1]
		}
		if strings.HasPrefix(sl, "data: ") {
			if event.Data != nil {
				c.recvdEvents = append(c.recvdEvents, event)
				c.c.Incr(1)
				event = Out{}
			}
			event.Data = line[len("data: ") : len(line)-1]
		}
		if event.Event != "" && event.Data != nil {
			c.recvdEvents = append(c.recvdEvents, event)
			c.c.Incr(1)
			event = Out{}
		}
	}
}
