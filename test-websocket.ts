#!/usr/bin/env ts-node

import WebSocket from 'ws';

// Type definitions for WebSocket message
interface WebSocketMessage {
  event: string;
  data: any;
  timestamp?: string;
  user_id?: number;
}

// User token (ID 24, role: user)
const token: string = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyNCwiZW1haWwiOiJ3c2FkbWluQGV4YW1wbGUuY29tIiwicm9sZSI6InVzZXIiLCJleHAiOjE3NjE5OTA2MDMsIm5iZiI6MTc2MTkwNDIwMywiaWF0IjoxNzYxOTA0MjAzfQ.6HpH3IMOViLfDkYYcuX4NPJ_O9y477Soh4OwrkBvO2I";

console.log("🔌 Connecting to WebSocket...");
const ws: WebSocket = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

ws.on('open', function open(): void {
  console.log('✅ WebSocket connected!');
});

ws.on('message', function incoming(data: Buffer | string): void {
  console.log('\n📨 Message received:');
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
  console.log('🔌 WebSocket disconnected');
  process.exit(0);
});

// Keep alive for 10 seconds
setTimeout((): void => {
  console.log('\n⏱️  Test timeout - closing connection');
  ws.close();
}, 10000);
