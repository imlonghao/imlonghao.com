package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	dbInit()
	r := gin.Default()
	r.Static("/static", "./static")
	r.StaticFile("/robots.txt", "./static/robots.txt")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")
	r.LoadHTMLGlob("views/*.html")
	r.GET("/", indexHandler)
	r.GET("/sitemap.xml", sitemapHandler)
	r.GET("/feed", feedHandler)
	r.GET("/rss", feedHandler)
	r.NoRoute(catchAllHandler)
	r.Run()
}
