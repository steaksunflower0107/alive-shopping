package main

import (
	"alivePlatform/order_service/config"
	"alivePlatform/order_service/controller"
	"alivePlatform/order_service/dao/mq"
	"alivePlatform/order_service/dao/mysql"
	"alivePlatform/order_service/dao/redis"
	"alivePlatform/order_service/logger"
	"alivePlatform/order_service/proto"
	"alivePlatform/order_service/registry"
	"alivePlatform/order_service/snowflake"

	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	// 加载配置文件
	var configFilePath string
	flag.StringVar(&configFilePath, "conf", "/Users/steaksunflower/Golang/src/alivePlatform/repertory_service/conf/config.yaml", "指定配置文件")
	flag.Parse()

	err := config.Init(configFilePath)
	if err != nil {
		zap.L().Panic("订单服务：加载配置文件失败")
		panic(err)
	}

	// 加载日志
	err = logger.Init(config.Conf.LogConfig, "dev")
	if err != nil {
		zap.L().Panic("订单服务：初始化日志失败")
		panic(err)
	}

	// 初始化Mysql
	err = mysql.Init(config.Conf.MySQLConfig)
	if err != nil {
		zap.L().Panic("订单服务：mysql初始化失败")
		panic(err)
	}

	// 初始化redis
	err = redis.Init(config.Conf.RedisConfig)
	if err != nil {
		zap.L().Panic("订单服务：redis初始化失败")
		panic(err)
	}

	// 初始化consul
	err = registry.Init(config.Conf.ConsulConfig.Addr)
	if err != nil {
		zap.L().Panic("库存服务：注册中心初始化失败")
		panic(err)
	}

	// 初始化snowflake
	err = snowflake.Init(config.Conf.StartTime, config.Conf.MachineID)
	if err != nil {
		panic(err)
	}

	// 初始化rocketmq
	err = mq.Init()
	if err != nil {
		panic(err)
	}
	// 监听订单超时的消息
	c, _ := rocketmq.NewPushConsumer(
		consumer.WithGroupName("order_srv_1"),
		consumer.WithNsResolver(primitive.NewPassthroughResolver([]string{"127.0.0.1:9876"})),
	)
	// 订阅topic
	err = c.Subscribe("xx_pay_timeout", consumer.MessageSelector{}, controller.OrderTimeoutHandle)
	if err != nil {
		fmt.Println(err.Error())
	}
	// Note: start after subscribe
	err = c.Start()
	if err != nil {
		panic(err)
	}

	lister, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Conf.RpcPort))
	if err != nil {
		zap.L().Panic("订单服务：RPC lister创建失败")
		panic(err)
	}

	// 创建grpc服务
	s := grpc.NewServer()
	// 注册健康检查，由此支持服务发现来进行健康检查
	grpc_health_v1.RegisterHealthServer(s, health.NewServer())
	// 注册RPC服务
	proto.RegisterRepertoryServer(s, &controller.OrderSrv{})
	// 启动服务
	go func() {
		err = s.Serve(lister)
		if err != nil {
			zap.L().Panic("启动order服务失败")
			panic(err)
		}
	}()

	zap.L().Info(fmt.Sprintf("rpc service start at %s:%d", config.Conf.IP, config.Conf.RpcPort))

	// 注册到consul
	err = registry.Reg.RegisterService(config.Conf.Name, config.Conf.IP, config.Conf.RpcPort, nil)
	if err != nil {
		zap.L().Panic("order服务注册到consul失败")
		return
	}
	zap.L().Info("order service start...")

	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	conn, err := grpc.DialContext( // RPC客户端
		context.Background(),
		fmt.Sprintf("%s:%d", config.Conf.IP, config.Conf.RpcPort),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()
	// Register Greeter
	err = proto.RegisterRepertoryHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	gwServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Conf.HttpPort),
		Handler: gwmux,
	}

	zap.L().Info("Serving gRPC-Gateway on http://0.0.0.0:" + strconv.Itoa(config.Conf.HttpPort))
	go func() {
		err := gwServer.ListenAndServe()
		if err != nil {
			log.Printf("gwServer.ListenAndServe failed, err: %v", err)
			return
		}
	}()

	// 服务退出时注销服务
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	serviceId := fmt.Sprintf("%s-%s-%d", config.Conf.Name, config.Conf.IP, config.Conf.RpcPort)
	registry.Reg.Deregister(serviceId)
}
