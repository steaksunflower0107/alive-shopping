package main

import (
	"alivePlatform/repertory_service/proto"
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	conn   *grpc.ClientConn
	client proto.RepertoryClient
)

func init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:8382", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	client = proto.NewRepertoryClient(conn)
}

func TestGetStock() {
	param := &proto.GoodsRepertoryInfo{
		GoodsId: 1,
		Num:     1,
	}

	resp, err := client.GetRepertory(context.Background(), param)
	fmt.Printf("resp:%v err:%v\n", resp, err)
}

func TestReduceStock(wg *sync.WaitGroup) {
	defer wg.Done()
	param := &proto.GoodsRepertoryInfo{
		GoodsId: 1,
		Num:     1,
	}
	resp, err := client.ReduceRepertory(context.Background(), param)
	fmt.Printf("resp:%v err:%v\n", resp, err)
}

func main() {
	defer conn.Close()
	//TestGetStock()
	var wg sync.WaitGroup
	for i := 0; i < 25; i++ {
		wg.Add(1)
		go TestReduceStock(&wg)
	}

	wg.Wait()
}
