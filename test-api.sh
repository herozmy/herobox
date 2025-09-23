#!/bin/bash

echo "测试内核路径检测 API..."
echo

# 启动服务器（后台）
cd /Users/herozmy/Desktop/herobox/bin/dev
./herobox &
SERVER_PID=$!

# 等待服务器启动
sleep 3

echo "调用路径检测 API..."
curl -s http://localhost:8080/api/singbox/kernel/detect-path | jq '.'

echo
echo "调用检查更新 API..."  
curl -s http://localhost:8080/api/singbox/kernel/check-update | jq '.'

# 停止服务器
kill $SERVER_PID

echo
echo "测试完成"
