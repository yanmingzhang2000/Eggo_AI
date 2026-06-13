#!/bin/bash
# 构建脚本
set -e
go build -o bin/server ./cmd/server
echo "Build complete: bin/server"
