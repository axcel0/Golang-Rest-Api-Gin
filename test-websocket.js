#!/usr/bin/env node

const WebSocket = require('ws');

// User token (ID 24, role: user)
const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyNCwiZW1haWwiOiJ3c2FkbWluQGV4YW1wbGUuY29tIiwicm9sZSI6InVzZXIiLCJleHAiOjE3NjE5OTA2MDMsIm5iZiI6MTc2MTkwNDIwMywiaWF0IjoxNzYxOTA0MjAzfQ.6HpH3IMOViLfDkYYcuX4NPJ_O9y477Soh4OwrkBvO2I";

console.log("🔌 Connecting to WebSocket...");
const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

ws.on('open', function open() {
  console.log('✅ WebSocket connected!');
});

ws.on('message', function incoming(data) {
  console.log('\n📨 Message received:');
  try {
    const message = JSON.parse(data);
    console.log(JSON.stringify(message, null, 2));
  } catch (e) {
    console.log(data.toString());
  }
});

ws.on('error', function error(err) {
  console.error('❌ WebSocket error:', err.message);
});

ws.on('close', function close() {
  console.log('🔌 WebSocket disconnected');
  process.exit(0);
});

// Keep alive for 10 seconds
setTimeout(() => {
  console.log('\n⏱️  Test timeout - closing connection');
  ws.close();
}, 10000);
