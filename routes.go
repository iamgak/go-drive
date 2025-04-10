package main

import (
	"time"

	"github.com/gin-gonic/gin"
)

func (app *Application) InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(MaintenanceMiddleware())
	r.Use(app.TimeoutMiddleware(5 * time.Second))
	// read API

	r.LoadHTMLGlob("templates/*.html")

	// Route for login page
	r.GET("/login", app.ShowLoginPage)
	r.GET("/register", app.ShowRegisterPage)
	authorise := r.Group("/drive")

	authorise.Use(app.LoginMiddleware(), secureHeaders(), app.rateLimiter())
	{
		// write API
		authorise.POST("/create", app.CreateFolder)
		authorise.POST("/upload/", app.UploadFile)
		authorise.PUT("/rename", app.RenameFolder)
		authorise.DELETE("/delete", app.DeleteFileOrFolder)
		authorise.GET("/*path", app.DriveListing)
	}

	r.POST("/login", app.UserLogin)
	r.POST("/register", app.UserRegister)
	r.GET("/activation_token/:token", app.UserActivateAccount)
	return r
}
