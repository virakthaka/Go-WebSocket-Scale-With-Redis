package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var redisClient *redis.Client

// Thread-safe in-memory map of connected clients
var clients = make(map[*websocket.Conn]bool)
var clientsMu sync.Mutex

// Message represents a chat message
type Message struct {
	Sender  string `json:"sender"`
	Content string `json:"content"`
}

func main() {
	// Initialize Redis
	redisClient = redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	// Fiber app setup
	app := fiber.New()

	// WebSocket upgrade route
	app.Get("/ws", websocket.New(handleWebSocket))

	// Start Redis subscriber goroutine
	go subscribeRedis()

	log.Println("WebSocket server running on :3000")
	log.Fatal(app.Listen(":3000"))
}

func handleWebSocket(c *websocket.Conn) {
	defer cleanupClient(c)

	clientsMu.Lock()
	clients[c] = true
	clientsMu.Unlock()

	for {
		var msg Message
		if err := c.ReadJSON(&msg); err != nil {
			log.Printf("Read error from %s: %v", c.RemoteAddr(), err)
			break
		}

		formatted := fmt.Sprintf("%s: %s", msg.Sender, msg.Content)

		// Publish to Redis
		if err := redisClient.Publish(ctx, "chat", formatted).Err(); err != nil {
			log.Printf("Redis publish error: %v", err)
		}
	}
}

// Subscribes to Redis "chat" channel and broadcasts to local clients
func subscribeRedis() {
	sub := redisClient.Subscribe(ctx, "chat")
	ch := sub.Channel()

	for msg := range ch {
		if msg == nil {
			log.Println("Received nil message from Redis")
			continue
		}

		broadcastMessage(msg.Payload)
	}
}

// Broadcasts a message to all connected WebSocket clients
func broadcastMessage(message string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for conn := range clients {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			log.Printf("Write error to %s: %v", conn.RemoteAddr(), err)
			if err = conn.Close(); err != nil {
				log.Printf("WebSocket close error: %v", err)
			}
			delete(clients, conn)
		}
	}
}

// cleanup all clients connected to the server
func cleanupClient(c *websocket.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	if _, ok := clients[c]; ok {
		delete(clients, c)
	}
	if err := c.Close(); err != nil {
		log.Printf("WebSocket close error: %v", err)
	}
}
