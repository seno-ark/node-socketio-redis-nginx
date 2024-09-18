const express = require('express');
const { createServer } = require('http');
const { Server } = require('socket.io');
const redis = require('redis');
const { createAdapter } = require('@socket.io/redis-adapter');

const redisURL = process.env.REDIS_URL;
const appName = process.env.NAME;

const pubClient = redis.createClient({ url: redisURL });
const subClient = pubClient.duplicate();

const app = express();
const httpServer = createServer(app);

const io = new Server(httpServer, {
  cors: {
    origin: "*",
    methods: ["GET", "POST"]
  },
});

Promise.all([pubClient.connect(), subClient.connect()]).then(() => {
  io.adapter(createAdapter(pubClient, subClient));

  io.on('connection', (socket) => {
    console.log(appName, 'New client connected');

    socket.on('join', (payload, ack_callback) => {
      console.log(appName, 'New Room Join', payload);

      const roomID = payload.room_id || "";
      if (!ack_callback) ack_callback = function(msg){}

      if (roomID.length > 0) {
        socket.join(roomID, (err, result) => {
            console.log(appName, 'join');
            if (err) console.log(err);
            else console.log(`Client joined room: ${payload}`);
        });
      } else ack_callback({'error': 'invalid_room'});
    });

    socket.on('disconnect', () => {
      console.log('Client disconnected');
    });
  });

  subClient.subscribe('socket.io', (message) => {
    console.log(`${appName} Got Message ${message}`);

    const data = JSON.parse(message);
    data.data[appName] = true
    if (data.room) {
      io.local.to(data.room).emit(data.event, data.data);
    } else {
      io.local.emit(data.event, data.data);
    }
  });

  const PORT = process.env.PORT || 3000;
  httpServer.listen(PORT, () => {
    console.log(`Socket.IO server running on port ${PORT}`);
  });
});