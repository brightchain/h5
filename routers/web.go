package routers

import (
	"h5/app/http/controllers"
	"h5/app/http/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterWebRouters(r *gin.Engine) {
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	i := new(controllers.Index)
	r.GET("/index", i.Index)
	dir := new(controllers.DirectoryClear)
	r.GET("/photo-clear", dir.PhotoDirClear)
	r.GET("/photo-month", dir.PhotoDirMonth)
	r.GET("/album-clear", dir.AlbumDirClear)
	r.GET("/calendar-clear", dir.CalendarDirClear)
	r.GET("/tshirt-clear", dir.TshirtDirClear)

	export := new(controllers.ExportExcel)
	r.GET("/hngx", export.Hngx)
	r.GET("/hnkj", export.Hnkj)
	r.GET("/wj", export.Smwj)
	exGroup := r.Group("/export")
	exGroup.Use(middleware.ExportExport())
	{
		exGroup.GET("/fjpa", export.Fjpa)
		exGroup.GET("/xinhua", export.Xinhua)
		exGroup.GET("/ydln", export.Ydln)
		exGroup.GET("/shtp", export.ShTp)
		exGroup.GET("/fjtp", export.FjTp)
		exGroup.GET("/nyorder", export.NyOrder)
		exGroup.GET("/hntborder", export.HntbOrder)
		exGroup.GET("/photocancel", export.PhotoCancal)
		exGroup.GET("/jxms", export.Jxms)
		exGroup.GET("/gdpa", export.GdpaOrder)
		exGroup.GET("/gdpazj", export.GdpaOrderZj)
		exGroup.GET("/gdpazs", export.GdpaOrderZs)
		exGroup.GET("/gdpas", export.GdpaOrders)
		exGroup.GET("/gdpaimport", export.GdpaImport)
		exGroup.GET("/yggxorder", export.YggxOrder)
		exGroup.GET("/hnms", export.Hnms)
		exGroup.GET("/gdtk", export.Gdtk)
		exGroup.GET("/gdtkorder", export.TkdgOrder)
		exGroup.GET("/fjrbs", export.FjrbsOrder)
		exGroup.GET("/gsqy", export.Gsqy)
		exGroup.GET("/gsqytotal", export.GsqyTotal)
		exGroup.GET("/xcgs", export.Xcgs)
		exGroup.GET("/whgs", export.Whgs)
	}

	aes := controllers.AesEcb{}
	r.POST("/aes/aes", aes.Aes)
	r.POST("/aes/encrypt", aes.Encrypt)
	r.POST("/aes/dow", aes.Down)
	r.GET("/rsa", aes.RsaDecrypt)
	car := controllers.Car{}
	r.GET("/car", car.Index)
	r.GET("/car_type", car.CarModel)
	r.GET("/car_detail", car.CarDetail)
	order := controllers.PayOrder{}
	orderGroup := r.Group("/order")
	orderGroup.Use(middleware.ExportExport())
	{
		orderGroup.GET("/product", order.GetOrderProduct)
		orderGroup.GET("/excelfix", order.ExcelFix)
		orderGroup.GET("/ddkf", order.Ddkf)
		orderGroup.GET("/zking", order.Zking)
		orderGroup.GET("/product-num", order.ProductNum)

	}

	down := new(controllers.DownOrder)
	r.GET("/mousedown", down.MouseOrderDown)
	sqlBak := new(controllers.Sqlbak)
	r.GET("/sqlbak", sqlBak.Index)
	r.GET("/sqlexcel", sqlBak.Excel)

	activity := new(controllers.Activity)
	r.GET("/activity", activity.UserReset)
	r.GET("/activity_cancel", activity.CancelOrder)

	mobile := new(controllers.MobileController)
	r.GET("/mobile", mobile.Get)

}
