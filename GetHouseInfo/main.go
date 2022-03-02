package main

import (
	"IHome/GetHouseInfo/handler"
	"IHome/GetHouseInfo/subscriber"

	"github.com/micro/go-grpc"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/util/log"

	GetHouseInfo "IHome/GetHouseInfo/proto/GetHouseInfo"
)

func main() {
	// New Service
	service := grpc.NewService(
		micro.Name("go.micro.srv.GetHouseInfo"),
		micro.Version("latest"),
	)

	// Initialise service
	service.Init()

	// Register Handler
	GetHouseInfo.RegisterGetHouseInfoHandler(service.Server(), new(handler.GetHouseInfo))

	// Register Struct as Subscriber
	micro.RegisterSubscriber("go.micro.srv.GetHouseInfo", service.Server(), new(subscriber.GetHouseInfo))

	// Register Function as Subscriber
	micro.RegisterSubscriber("go.micro.srv.GetHouseInfo", service.Server(), subscriber.Handler)

	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
