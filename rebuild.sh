#!/bin/bash

echo "Pull code change"
git pull

# 日志文件路径
LOG_FILE="backup-go.log"

# 查找 backup-go 进程 ID
PID=$(lsof -t -c backup-go)

# 如果进程存在
if [ -n "$PID" ]; then
  # 杀死进程
  kill -9 $PID
  echo "Killed backup-go process with PID $PID"
else
  echo "No backup-go process found"
fi

# 编译新的二进制文件
go build -o backup-go
echo "Compiled new backup-go binary"

# 启动新的进程,将输出重定向到日志文件
nohup ./backup-go >> "$LOG_FILE" 2>&1 &
echo "Started new backup-go process. Logs will be written to $LOG_FILE"