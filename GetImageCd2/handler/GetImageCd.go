package handler

import (
	example "IHome/GetImageCd2/proto/GetImageCd2"
	"IHome/IhomeWeb/utils"
	"context"
	"image/color"
	"time"

	"github.com/afocus/captcha"
)

type GetImageCd2 struct{}

// Call is a single request handler called via client.Call or the generated client code
func (e *GetImageCd2) GetImageCd(ctx context.Context, req *example.Request, rsp *example.Response) error {
	//1 初始化返回值
	rsp.Error = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Error)

	//2 生成随即数与图片
	//创建一个句柄

	cap := captcha.New()
	//设置字体
	if err := cap.SetFont("GetImageCd2/comic.ttf"); err != nil {
		//抛出异常
		panic(err.Error())
	}

	//设置突破大小
	cap.SetSize(90, 41)
	//设置干扰强度
	cap.SetDisturbance(captcha.MEDIUM)
	//设置前景色
	cap.SetFrontColor(color.RGBA{255, 255, 255, 255}, color.RGBA{0, 0, 0, 200})
	//设置背景色
	cap.SetBkgColor(color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 255, 255}, color.RGBA{0, 153, 0, 255})

	//随即生存 图片 与 验证码
	img, str := cap.Create(4, captcha.NUM)

	//打印字符串
	println(str)

	//3 获取uuid
	uuid := req.Uuid

	//4 连接redis
	bm, err := utils.RedisOpen(utils.G_server_name, utils.G_redis_addr,
		utils.G_redis_port, utils.G_redis_dbnum)
	if err != nil {
		rsp.Error = utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Error)
		return nil
	}

	//5 存入数据
	bm.Put(uuid, str, time.Second*300)

	//6 将图片拆分 赋值到proto
	a := *img

	b := *(a.RGBA)

	//pix
	for _, value := range b.Pix {

		rsp.Pix = append(rsp.Pix, uint32(value))
	}

	//stride
	rsp.Stride = int64(b.Stride)

	//point
	rsp.Min = &example.ResponsePoint{X: int64(b.Rect.Min.X),
		Y: int64(b.Rect.Min.Y)}
	rsp.Max = &example.ResponsePoint{X: int64(b.Rect.Max.X),
		Y: int64(b.Rect.Max.Y)}

	return nil
}
