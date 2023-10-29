package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"mxshop_srvs/goods_srv/proto"
)

var brandClient proto.GoodsClient
var conn *grpc.ClientConn

func Init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		panic("连接失败")
	}

	brandClient = proto.NewGoodsClient(conn)
}

func TestGetBrand() {
	rsp, err := brandClient.BrandList(context.Background(), &proto.BrandFilterRequest{})
	if err != nil {
		panic("查询失败")
	}
	for _, brand := range rsp.Data {
		fmt.Println(brand.Name)
	}
}

func main() {
	Init()
	TestGetBrand()
	conn.Close()
}
