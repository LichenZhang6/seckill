package main

import (
	"context"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/kataras/iris/v12/sessions"
	"github.com/opentracing/opentracing-go/log"
	"seckill/common"
	"seckill/frontend/middleware"
	"seckill/frontend/web/controllers"
	"seckill/repositories"
	"seckill/services"
	"time"
)

func main() {
	// 1.创建iris实例
	app := iris.New()

	// 2.设置错误模式，在MVC模式下提示错误
	app.Logger().SetLevel("debug")

	// 3.注册模板
	template := iris.HTML("./frontend/web/views", ".html").Layout("shared/layout.html").Reload(true)
	app.RegisterView(template)

	// 4.设置模板目录
	app.HandleDir("/public", "./frontend/web/public")

	// 访问生成好的html静态文件
	app.HandleDir("/html", "./frontend/web/htmlProductShow")

	// 出现异常跳转到指定页面
	app.OnAnyErrorCode(func(ctx iris.Context) {
		ctx.ViewData("message", ctx.Values().GetStringDefault("message", "访问页面出错！"))
		ctx.ViewLayout("")
		ctx.View("shared/error.html")
	})

	// 连接数据库
	db, err := common.NewMysqlConn()
	if err != nil {
		log.Error(err)
	}

	session := sessions.New(sessions.Config{
		Cookie:  "AdminCookie",
		Expires: 60 * time.Minute,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 5.注册控制器
	userRepository := repositories.NewUserManager("user", db)
	userService := services.NewUserService(userRepository)
	userParty := app.Party("/user")
	user := mvc.New(userParty)
	user.Register(ctx, userService, session.Start)
	user.Handle(new(controllers.UserController))

	product := repositories.NewProductManager("product", db)
	productService := services.NewProductService(product)
	order := repositories.NewOrderManager("order", db)
	orderService := services.NewOrderService(order)
	productParty := app.Party("/product")
	pro := mvc.New(productParty)
	productParty.Use(middleware.AuthConProduct)
	pro.Register(productService, orderService, session.Start)
	pro.Handle(new(controllers.ProductController))

	app.Run(
		iris.Addr("localhost:8082"),
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithOptimizations,
	)
}
