#!/usr/bin/env node

const WebSocket = require('ws');
const https = require('http');

// User token (ID 24, role: user)
const token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoyNCwiZW1haWwiOiJ3c2FkbWluQGV4YW1wbGUuY29tIiwicm9sZSI6InVzZXIiLCJleHAiOjE3NjE5OTA2MDMsIm5iZiI6MTc2MTkwNDIwMywiaWF0IjoxNzYxOTA0MjAzfQ.6HpH3IMOViLfDkYYcuX4NPJ_O9y477Soh4OwrkBvO2I";

let messageCount = 0;

console.log("🔌 Connecting to WebSocket...");
const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

ws.on('open', function open() {
  console.log('✅ WebSocket connected!\n');
  
  // Wait 1 second then check stats
  setTimeout(() => {
    console.log('📊 Checking WebSocket stats...');
    const options = {
      hostname: 'localhost',
      port: 8080,
      path: '/ws/stats',
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    };
    
    const req = https.request(options, (res) => {
      let data = '';
      res.on('data', (chunk) => data += chunk);
      res.on('end', () => {
        console.log('Stats response:', data, '\n');
        
        // Now try to broadcast (will fail - user not admin)
        console.log('📡 Attempting broadcast (should fail - not admin)...');
        const broadcastOptions = {
          hostname: 'localhost',
          port: 8080,
          path: '/ws/broadcast',
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
          }
        };
        
        const broadcastReq = https.request(broadcastOptions, (res) => {
          let data = '';
          res.on('data', (chunk) => data += chunk);
          res.on('end', () => {
            console.log('Broadcast response:', data, '\n');
            
            // Close after 2 more seconds
            setTimeout(() => {
              console.log('⏱️  Test complete - closing connection');
              ws.close();
            }, 2000);
          });
        });
        
        broadcastReq.write(JSON.stringify({
          message: "Test broadcast from Node.js",
          priority: "high"
        }));
        broadcastReq.end();
      });
    });
    
    req.end();
  }, 1000);
});

ws.on('message', function incoming(data) {
  messageCount++;
  console.log(`\n📨 Message #${messageCount}:`);
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
  console.log('\n🔌 WebSocket disconnected');
  console.log(`📊 Total messages received: ${messageCount}`);
  process.exit(0);
});
