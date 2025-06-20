package main

import (
	"context"
	"log"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var redisClient *redis.Client

// Thread-safe in-memory map of connected clients
var clients = make(map[*websocket.Conn]string)
var clientsMu sync.Mutex

// Track which rooms we've already subscribed to
var subscribedRooms = make(map[string]bool)
var subscribedRoomsMu sync.Mutex

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
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{Views: engine})

	app.Get("/chat/:room", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{"Room": c.Params("room")})
	})

	// WebSocket upgrade route
	app.Get("/ws/:room", websocket.New(handleWebSocket))

	log.Println("WebSocket server running on :3000")
	log.Fatal(app.Listen(":3000"))
}

func handleWebSocket(c *websocket.Conn) {
	room := c.Params("room")
	go subscribeRedis(room)
	defer cleanupClient(c)

	clientsMu.Lock()
	clients[c] = room
	clientsMu.Unlock()

	for {
		var msg Message
		if err := c.ReadJSON(&msg); err != nil {
			log.Printf("Read error from %s: %v", c.RemoteAddr(), err)
			break
		}

		if data, err := sonic.Marshal(msg); err == nil {
			if err = redisClient.Publish(ctx, "chat:"+room, data).Err(); err != nil {
				log.Printf("Redis publish error: %v", err)
			}
		}
	}
}

// Subscribes to Redis "chat" channel and broadcasts to local clients
func subscribeRedis(room string) {
	subscribedRoomsMu.Lock()
	defer subscribedRoomsMu.Unlock()

	if subscribedRooms[room] {
		return
	}
	subscribedRooms[room] = true
	sub := redisClient.Subscribe(ctx, "chat:"+room)

	for msg := range sub.Channel() {
		broadcastMessage(room, []byte(msg.Payload))
	}
}

// Broadcasts a message to all connected WebSocket clients
func broadcastMessage(room string, message []byte) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for c, r := range clients {
		if r == room {
			if err := c.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Write error to %s: %v", c.RemoteAddr(), err)
				delete(clients, c)
				_ = c.Close()
			}
		}
	}
}

// cleanup all clients connected to the server
func cleanupClient(c *websocket.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	delete(clients, c)
	_ = c.Close()
}
