package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var redisClient *redis.Client

// In-memory map of connected clients
var clients = make(map[*websocket.Conn]bool)

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
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		defer func() {
			delete(clients, c)
			c.Close()
		}()
		clients[c] = true

		for {
			var msg Message
			if err := c.ReadJSON(&msg); err != nil {
				log.Println("read error:", err)
				break
			}

			// Publish to Redis
			if err := redisClient.Publish(ctx, "chat", fmt.Sprintf("%s: %s", msg.Sender, msg.Content)).Err(); err != nil {
				log.Println("publish error:", err)
			}
		}
	}))

	// Start Redis subscriber goroutine
	go subscribeRedis()

	log.Println("WebSocket server running on :3000")
	log.Fatal(app.Listen(":3000"))
}

// Subscribes to Redis "chat" channel and broadcasts to local clients
func subscribeRedis() {
	sub := redisClient.Subscribe(ctx, "chat")
	ch := sub.Channel()

	for msg := range ch {
		broadcastMessage(msg.Payload)
	}
}

// Broadcasts a message to all connected WebSocket clients
func broadcastMessage(message string) {
	for conn := range clients {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			log.Println("write error:", err)
			conn.Close()
			delete(clients, conn)
		}
	}
}
