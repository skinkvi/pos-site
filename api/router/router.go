package routers

import (
	handler "Positiv/api/handler"
	"Positiv/internal/middleware"
	"os"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	// Типо как рабочий каталог
	workingDir, _ := os.Getwd()

	router.LoadHTMLGlob(workingDir + "/templates/*")
	router.Static("/static", "./static/")
	router.Static("/image", "./image")

	router.GET("/services/:name", handler.ServicesHandler)
	router.GET("/:name", handler.DesignDevelopmentHandler)
	router.GET("/", handler.MainHandler)
	router.GET("/contacts", handler.GetContact)

	router.GET("/admin/login", handler.LoginHandler)
	router.POST("/admin/login", handler.LoginPostHandler)

	admin := router.Group("/admin")
	admin.Use(middleware.AdminAuthMiddleware())
	{
		admin.GET("/", handler.AdminHandlerMain)
		admin.GET("/create_page", handler.CreatePageHandler)
		admin.GET("/edit_page/:id", handler.EditPageHandler)
		admin.POST("/saveUpdatePage", handler.UpdatePageHandler)
	}
	router.POST("/admin/save_page", handler.SavePageHandler)

	return router
}
