package subscriber

import (
	example "IHome/GetImageCd2/proto/GetImageCd2"
	"context"

	"github.com/micro/go-micro/util/log"
)

type GetImageCd2 struct{}

func (e *GetImageCd2) Handle(ctx context.Context, msg *example.Request) error {
	log.Log("Handler Received message: ", msg)
	return nil
}

func Handler(ctx context.Context, msg *example.Request) error {
	log.Log("Function Received message: ", msg)
	return nil
}
