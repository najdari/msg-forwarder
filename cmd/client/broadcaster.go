package main

import (
	"context"
	"fmt"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) publish(msg string) {
	h.broadcast <- []byte(msg)
}

func (h *Hub) run(ctx context.Context) {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				case <-ctx.Done():
					fmt.Println("cancelling message broadcast")
					return
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		case <-ctx.Done():
			fmt.Println("shutting down broadcaster hub")
			return
		}
	}
}
