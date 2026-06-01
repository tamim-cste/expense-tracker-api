// Package main is the entry point for the Expense Tracker API.

package main

import (
	_ "expense-tracker-api/routers"

	"github.com/beego/beego/v2/server/web/filter/cors"
	beego "github.com/beego/beego/v2/server/web"
)

func main() {
	
	beego.BConfig.WebConfig.AutoRender = false

	// This will Enable Swagger UI in dev mode
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	// Enable CORS so any frontend origin can connect
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "X-User-ID", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	beego.Run()
}
