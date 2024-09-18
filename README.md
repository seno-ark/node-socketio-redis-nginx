# node-socketio-redis-nginx

Socket.IO Implementation with Redis PubSub and Nginx.

## Run with Docker Compose

docker-compose up

## Access the Application

http://localhost?room_id=123

## Broadcast to a Room via API

curl -XPOST localhost/api/emit -d '{"room": "123", "content":"hi room 123"}'
