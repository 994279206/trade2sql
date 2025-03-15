#!/bin/bash

# 确保资源目录存在
mkdir -p /Users/suhualin/Desktop/work/trade2sql/assets

# 如果图标文件不存在，创建一个简单的图标
if [ ! -f /Users/suhualin/Desktop/work/trade2sql/assets/icon.png ]; then
    echo "图标文件不存在，请将图标文件放置在 assets/icon.png"
    exit 1
fi

# 生成资源包
cd /Users/suhualin/Desktop/work/trade2sql
fyne bundle -o internal/resources/bundled.go -name AppIcon assets/icon.png