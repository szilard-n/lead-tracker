package handler

import (
	"fmt"
	"leadgentracker/internals/model/constants"
	"log"
	"net/http"
	"sync"
)

type Client struct {
	channel chan string
}

type SSEBroadcaster struct {
	clients map[*Client]bool
	mutex   sync.Mutex
}

func NewSSEBroadcaster() *SSEBroadcaster {
	return &SSEBroadcaster{
		clients: make(map[*Client]bool),
	}
}

// HandleSSE handles new browser connections
func (b *SSEBroadcaster) HandleSSE(w http.ResponseWriter, r *http.Request) {
	// set up headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// create a new client with a message channel
	client := &Client{
		channel: make(chan string, 10),
	}

	// register client
	b.mutex.Lock()
	b.clients[client] = true
	b.mutex.Unlock()

	// remove connection when client closes
	defer func() {
		b.mutex.Lock()
		delete(b.clients, client)
		close(client.channel)
		b.mutex.Unlock()
	}()

	// keep connection open and send messages as they come
	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Println("streaming not supported")
		http.Error(w, constants.ErrorMessage, http.StatusInternalServerError)
		return
	}

	log.Println("created SSE connection")

	for {
		select {
		case msg := <-client.channel:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func (b *SSEBroadcaster) Broadcast(message string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	log.Printf("Broadcasting message: %s", message)

	for client := range b.clients {
		select {
		case client.channel <- message:
			log.Printf("Message sent successfully to client")
		default:
			log.Printf("Failed to send message to client - removing client")
			delete(b.clients, client)
			close(client.channel)
		}
	}
}
