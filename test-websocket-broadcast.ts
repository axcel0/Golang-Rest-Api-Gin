#!/usr/bin/env ts-node

import WebSocket from 'ws';
import http from 'http';

// Type definitions
interface WebSocketMessage {
  event: string;
  data: any;
  timestamp?: string;
  user_id?: number;
}

interface BroadcastRequest {
  message: string;
  priority: string;
}

// User token (ID 24, role: user)
const token: string = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyNCwiZW1haWwiOiJ3c2FkbWluQGV4YW1wbGUuY29tIiwicm9sZSI6InVzZXIiLCJleHAiOjE3NjE5OTA2MDMsIm5iZiI6MTc2MTkwNDIwMywiaWF0IjoxNzYxOTA0MjAzfQ.6HpH3IMOViLfDkYYcuX4NPJ_O9y477Soh4OwrkBvO2I";

let messageCount: number = 0;

console.log("🔌 Connecting to WebSocket...");
const ws: WebSocket = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

ws.on('open', function open(): void {
  console.log('✅ WebSocket connected!\n');
  
  // Wait 1 second then check stats
  setTimeout((): void => {
    console.log('📊 Checking WebSocket stats...');
    const options: http.RequestOptions = {
      hostname: 'localhost',
      port: 8080,
      path: '/ws/stats',
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    };
    
    const req = http.request(options, (res: http.IncomingMessage): void => {
      let data: string = '';
      res.on('data', (chunk: Buffer): void => {
        data += chunk.toString();
      });
      res.on('end', (): void => {
        console.log('Stats response:', data, '\n');
        
        // Now try to broadcast (will fail - user not admin)
        console.log('📡 Attempting broadcast (should fail - not admin)...');
        const broadcastOptions: http.RequestOptions = {
          hostname: 'localhost',
          port: 8080,
          path: '/ws/broadcast',
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        };
        
        const broadcastReq = http.request(broadcastOptions, (res: http.IncomingMessage): void => {
          let data: string = '';
          res.on('data', (chunk: Buffer): void => {
            data += chunk.toString();
          });
          res.on('end', (): void => {
            console.log('Broadcast response:', data, '\n');
            
            // Close after 2 more seconds
            setTimeout((): void => {
              console.log('⏱️  Test complete - closing connection');
              ws.close();
            }, 2000);
          });
        });
        
        const broadcastData: BroadcastRequest = {
          message: "Test broadcast from TypeScript",
          priority: "high"
        };
        
        broadcastReq.write(JSON.stringify(broadcastData));
        broadcastReq.end();
      });
    });
    
    req.end();
  }, 1000);
});

ws.on('message', function incoming(data: Buffer | string): void {
  messageCount++;
  console.log(`\n📨 Message #${messageCount}:`);
  try {
    const dataString: string = typeof data === 'string' ? data : data.toString();
    const message: WebSocketMessage = JSON.parse(dataString);
    console.log(JSON.stringify(message, null, 2));
  } catch (e) {
    const dataString: string = typeof data === 'string' ? data : data.toString();
    console.log(dataString);
  }
});

ws.on('error', function error(err: Error): void {
  console.error('❌ WebSocket error:', err.message);
});

ws.on('close', function close(): void {
  console.log('\n🔌 WebSocket disconnected');
  console.log(`📊 Total messages received: ${messageCount}`);
  process.exit(0);
});
