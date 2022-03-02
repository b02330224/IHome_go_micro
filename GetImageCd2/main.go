package main

import (
	"IHome/GetImageCd2/handler"
	"IHome/GetImageCd2/subscriber"

	"github.com/micro/go-grpc"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/util/log"

	GetImageCd2 "IHome/GetImageCd2/proto/GetImageCd2"
)

func main() {
	// New Service
	service := grpc.NewService(
		micro.Name("go.micro.srv.GetImageCd2"),
		micro.Version("latest"),
	)

	// Initialise service
	service.Init()

	// Register Handler
	GetImageCd2.RegisterGetImageCd2Handler(service.Server(), new(handler.GetImageCd2))

	// Register Struct as Subscriber
	micro.RegisterSubscriber("go.micro.srv.GetImageCd2", service.Server(), new(subscriber.GetImageCd2))

	// Register Function as Subscriber
	micro.RegisterSubscriber("go.micro.srv.GetImageCd2", service.Server(), subscriber.Handler)

	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
