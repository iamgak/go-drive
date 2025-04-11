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

	authorise := r.Group("/drive")

	authorise.Use(app.LoginMiddleware(), secureHeaders(), app.rateLimiter())
	{
		//listing of all the users files and folders
		authorise.GET("/*path", app.DriveListing)
		// write API
		authorise.POST("/create", app.CreateFolder)         //Create new folder
		authorise.POST("/upload/", app.UploadFile)          //Create new file
		authorise.PUT("/rename", app.RenameFolder)          // rename file or folder
		authorise.DELETE("/delete", app.DeleteFileOrFolder) // deleter file or folder
	}

	//html pages
	r.GET("/login", app.ShowLoginPage)
	r.GET("/register", app.ShowRegisterPage)

	// req handle
	r.POST("/login", app.UserLogin)
	r.POST("/register", app.UserRegister)

	//account activate after registration
	r.GET("/activation_token/:token", app.UserActivateAccount)
	return r
}
