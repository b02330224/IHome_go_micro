package handler

import (
	example "IHome/GetUserHouses/proto/example"
	"IHome/IhomeWeb/models"
	"IHome/IhomeWeb/utils"
	"context"
	"encoding/json"
	"log"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/garyburd/redigo/redis"
)

type Example struct{}

func (e *Example) GetUserHouses(ctx context.Context, req *example.Request, rsp *example.Response) error {
	beego.Info("获取当前用户所发布的房源 GetUserHouses /api/v1.0/user/houses")
	rsp.Errno = utils.RECODE_OK
	rsp.Errmsg = utils.RecodeText(rsp.Errno)
	//连接redis
	bm, err := utils.RedisOpen(utils.G_server_name, utils.G_redis_addr, utils.G_redis_port, utils.G_redis_dbnum)
	if err != nil {
		log.Println("redis connect err")
		rsp.Errno = utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}
	value := bm.Get(req.Sessionid + "user_id")
	user_id, _ := redis.Int(value, nil)
	o := orm.NewOrm()
	var houses []models.House
	num, err := o.QueryTable("house").Filter("user__id", user_id).All(&houses)
	if err != nil {
		rsp.Errno = utils.RECODE_DBERR
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil
	}
	if num == 0 {
		rsp.Errno = utils.RECODE_NODATA
		rsp.Errmsg = utils.RecodeText(rsp.Errno)
		return nil

	}
	//转换为二进制传输给前端
	house, err := json.Marshal(houses)
	rsp.Mix = house
	return nil
}
