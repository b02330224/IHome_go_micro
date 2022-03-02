package subscriber

import (
	"context"

	"github.com/micro/go-micro/util/log"

	example "IHome/GetHouseInfo/proto/GetHouseInfo"
)

type GetHouseInfo struct{}

func (e *GetHouseInfo) Handle(ctx context.Context, msg *example.Message) error {
	log.Log("Handler Received message: ", msg.Say)
	return nil
}

func Handler(ctx context.Context, msg *example.Message) error {
	log.Log("Function Received message: ", msg.Say)
	return nil
}
