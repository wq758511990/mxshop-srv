package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"mxshop_srvs/user_srv/proto"
)

var userClient proto.UserClient
var conn *grpc.ClientConn

func Init() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		panic("连接失败")
	}

	userClient = proto.NewUserClient(conn)
}

func TestGetUserList() {
	rsp, err := userClient.GetUserList(context.Background(), &proto.PageInfo{
		Pn:    1,
		PSize: 5,
	})
	if err != nil {
		panic("查询失败")
	}
	for _, user := range rsp.Data {
		fmt.Println(user.Mobile, user.NickName, user.Password)
		checkResult, _ := userClient.CheckPassword(context.Background(), &proto.PasswordCheckInfo{
			Password:          "admin123",
			EncryptedPassword: user.Password,
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(checkResult)
	}
}

func TestCreateUser() {
	for i := 0; i < 10; i++ {
		rsp, err := userClient.CreateUser(context.Background(), &proto.CreateUserInfo{
			Password: "admin123",
			Mobile:   fmt.Sprintf("18766665555-%d", i),
			NickName: fmt.Sprintf("Bobby-%d", i),
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(rsp.Id)
	}
}

func main() {
	Init()
	TestGetUserList()
	//TestCreateUser()
	conn.Close()
}
