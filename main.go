package main

import (
	"fmt"
	"log"
	"os"
	"projectK/conf"
	"projectK/network"

	"github.com/bwmarrin/snowflake"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

var g_signal = make(chan os.Signal, 1)

func initLogger() {

}

func main() {
	// 配置
	config, err := conf.ReadConfig()
	if err != nil {
		log.Fatalf("无法读取配置文件：%v", err)
	}

	// 日志
	logFile, err := os.OpenFile(config.LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("无法创建日志文件")
		return
	}

	defer logFile.Close()
	logger := logrus.New()
	logger.SetOutput(logFile)
	//snowflake
	node, err := snowflake.NewNode(1)
	if err != nil {
		logger.Println("Snowflake node error...")
		return
	}
	// 服务器启动
	s := network.NewServer(logger, node, redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 0}))
	s.Serve()

}
