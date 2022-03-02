package main

import (
	"IHome/IhomeWeb/handler"
	_ "IHome/IhomeWeb/models"

	"github.com/julienschmidt/httprouter"
	"github.com/micro/go-log"
	"github.com/micro/go-web"

	"net/http"
)

func main() {
	// 构造web服务
	service := web.NewService(
		web.Name("go.micro.web.IhomeWeb"),
		web.Version("latest"),
		web.Address(":10086"),
	)
	// 服务初始化
	if err := service.Init(); err != nil {
		log.Fatal(err)
	}
	//构建路由
	rou := httprouter.New()

	//将路由注册到服务
	rou.NotFound = http.FileServer(http.Dir("./IhomeWeb/static"))
	//rou.ServeFiles("/static/*filepath", http.Dir("static"))
	rou.GET("/api/v1.0/areas", handler.GetArea)
	//欺骗浏览器  session index
	rou.GET("/api/v1.0/session", handler.GetSession)

	// //获取图片验证码
	rou.GET("/api/v1.0/imagecode/:uuid", handler.GetImageCd)
	//获取图片验证码2
	//rou.GET("/api/v1.0/imagecode/:uuid", handler.GetImageCd2)
	//获取短信验证码
	rou.GET("/api/v1.0/smscode/:mobile", handler.Getsmscd)
	//用户注册
	rou.POST("/api/v1.0/users", handler.PostRet)
	//用户登陆
	rou.POST("/api/v1.0/sessions", handler.PostLogin)
	//退出登陆
	rou.DELETE("/api/v1.0/session", handler.DeleteSession)
	// //获取用户详细信息
	rou.GET("/api/v1.0/user", handler.GetUserInfo)
	// //用户上传图片
	rou.POST("/api/v1.0/user/avatar", handler.PostAvatar)
	// //请求更新用户名
	rou.PUT("/api/v1.0/user/name", handler.PutUserInfo)
	//身份认证检查 同  获取用户信息   所调用的服务是 GetUserInfo
	rou.GET("/api/v1.0/user/auth", handler.GetUserAuth)
	//实名认证服务
	rou.POST("/api/v1.0/user/auth", handler.PostUserAuth)
	//获取用户已发布房源信息服务
	rou.GET("/api/v1.0/user/houses", handler.GetUserHouses)
	//发送（发布）房源信息服务
	rou.POST("/api/v1.0/houses", handler.PostHouses)
	//发送（上传）房屋图片服务
	rou.POST("/api/v1.0/houses/:id/images", handler.PostHouseImage)
	//获取房屋详细信息的服务
	rou.GET("/api/v1.0/houses/:id", handler.GetHouseInfo)
	//获取首页轮播图的服务
	rou.GET("/api/v1.0/house/index", handler.GetIndex)
	//搜索房源服务
	rou.GET("/api/v1.0/houses", handler.GetHouses)
	//发布订单服务
	rou.POST("/api/v1.0/orders", handler.PostOrders)
	//获取房东/租户订单信息服务
	rou.GET("/api/v1.0/user/orders", handler.GetUserOrder)
	//更新房东同意/拒绝订单
	rou.PUT("/api/v1.0/orders/:id/status", handler.PutOrders)
	//更新用户评价订单信息
	rou.PUT("/api/v1.0/orders/:id/comment", handler.PutComment)
	service.Handle("/", rou)
	//service.Handle("/", http.FileServer(http.Dir("html")))
	// 服务运行
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
