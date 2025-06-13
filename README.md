# Scalable WebSocket Server With Redis

This project demonstrates how to scale **Golang** WebSocket servers horizontally using:

- 🧠 **Go + Fiber + Fiber WebSocket** for services backbone
- 🔁 **Redis Pub/Sub** for cross-instance communication
- 🔀 **NGINX** as a reverse proxy and load balancer
- 🐳 **Docker Compose** to run services locally

## 🚀 Project Purpose

WebSocket servers maintain persistent client connections, but when scaled horizontally, each instance handles its own isolated connections. This creates a challenge: messages can't reach clients connected to other instances:

> ❗ Messages sent from **Instance A** will not reach clients connected to **Instance B**

This project solves that problem using **Redis Pub/Sub**, which allows different server instances to communicate and broadcast messages across all connected clients, regardless of which server they are connected to.

The project demonstrates:

- How to **scale WebSocket servers horizontally**
- How to use **Redis for message broadcasting between instances**
- How to implement a **load-balanced WebSocket endpoint using NGINX**

## Features

- WebSocket server with `Json` message handling
- Load balancing with `Nginx`
- Cross-instance message broadcasting via Redis `Pub/Sub`
- Easily extensible for authentication, channels, presence, etc.

## Requirements

- Docker + Docker Compose
- Go (only for local binary build)

## Build Instructions

### 1. Clone the Repository

```bash
git clone https://github.com/virakthaka/go-websocket-scale-with-redis.git
cd go-websocket-scale-with-redis
```

### 2. Build the Go App Binary

```bash
GOOS=linux GOARCH=amd64 go build -o bin/app
```

> This generates a binary at `bin/app` that will be used inside Docker containers.

## 🐳 Run the Services in Docker

```bash
docker-compose up -d --build
```

> Action command:
> - Build App binary image
> - Start Redis and Nginx Server
> - Start two app instances (`app1`, `app2`)
> - Start Endpoint `ws://localhost:8080/ws`

## Structure

```
src/
├── main.go              # Go WebSocket server
│── go.mod
│── bin/app              # Compiled Go binary (built manually)
├── nginx/
│   └── default.conf     # NGINX config for load balancing
├── docker-compose.yml
├── Dockerfile           # For docker image Go service
```

## Testing

### Option 1: Using `Postman`

Connect two postman tabs:

```bash
ws://localhost:8080/ws
```

Send a message in one, and you'll receive it in both!

```json
{"sender":"Alice","content":"Hello everyone!"}
```

### Option 2: Browser Client

Create a simple `index.html` and open it in two tabs:

```js
<input id="input" placeholder="Type a message..." />
<button onclick="sendMsg()">Send</button>
<pre id="log"></pre>
<script>
  const ws = new WebSocket("ws://localhost:8080/ws");

  ws.onopen = () => log("Connected");
  ws.onclose = () => log("Disconnected");
  ws.onmessage = (msg) => log(msg.data);

  function sendMsg() {
    const msg = document.getElementById("input").value;
    const json = JSON.stringify({ sender: "Blob", content: msg });
    ws.send(json);
  }

  function log(msg) {
    document.getElementById("log").textContent += msg + "\n";
  }
</script>
```

## License

This project is open-source and available under the [MIT License]().
