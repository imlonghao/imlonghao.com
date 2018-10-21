package main

import (
	"fmt"
	"html/template"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
	"github.com/gorilla/feeds"
)

func indexHandler(c *gin.Context) {
	var articles []indexDict
	var links []linkDict
	rows, _ := pool.Query("getArticles")
	for rows.Next() {
		var id int
		var title string
		_ = rows.Scan(&id, &title)
		article := indexDict{id, title}
		articles = append(articles, article)
	}
	rows, _ = pool.Query("getLinks")
	for rows.Next() {
		var url string
		var name string
		_ = rows.Scan(&url, &name)
		link := linkDict{url, name}
		links = append(links, link)
	}
	c.HTML(200, "index.html", gin.H{
		"title":       "首页",
		"description": "非专业人士的非专业博客",
		"articles":    articles,
		"links":       links,
	})
}

func articleHandler(c *gin.Context, articleID string) {
	var id int
	var title string
	var description string
	var content []byte
	_ = pool.QueryRow("getArticle", articleID).Scan(&id, &title, &description, &content)
	if id == 0 {
		c.JSON(404, gin.H{"error": "404"})
		return
	}
	content = markdown.ToHTML(content, nil, nil)
	c.HTML(200, "article.html", gin.H{
		"id":          id,
		"title":       title,
		"description": description,
		"content":     template.HTML(string(content)),
	})
}

func sitemapHandler(c *gin.Context) {
	sitemap := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.sitemaps.org/schemas/sitemap/0.9 http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd" xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	<url>
		<loc>https://imlonghao.com/</loc>
		<lastmod>%s</lastmod>
		<changefreq>daily</changefreq>
		<priority>1.0</priority>
	</url>
`, time.Now().Format(time.RFC3339))
	rows, _ := pool.Query("getSitemap")
	for rows.Next() {
		var id int
		var createTime string
		_ = rows.Scan(&id, &createTime)
		s := fmt.Sprintf(`    <url>
        <loc>https://imlonghao.com/%d.html</loc>
        <lastmod>%s</lastmod>
        <changefreq>monthly</changefreq>
        <priority>0.5</priority>
	</url>
`, id, createTime)
		sitemap += s
	}
	sitemap += `</urlset>`
	c.Data(200, "text/xml", []byte(sitemap))
}

func feedHandler(c *gin.Context) {
	feed := feeds.Feed{
		Title:       "imlonghao",
		Link:        &feeds.Link{Href: "https://imlonghao.com/"},
		Description: "非专业人士的非专业博客",
		Created:     time.Now(),
	}
	rows, _ := pool.Query("getFeed")
	for rows.Next() {
		var id int
		var title string
		var description string
		var createTime time.Time
		_ = rows.Scan(&id, &title, &description, &createTime)
		url := fmt.Sprintf("https://imlonghao.com/%d.html", id)
		feed.Items = append(feed.Items, &feeds.Item{
			Title:       title,
			Link:        &feeds.Link{Href: url},
			Description: description,
			Created:     createTime,
		})
	}
	rss, _ := feed.ToRss()
	c.Data(200, "text/xml", []byte(rss))
}

func catchAllHandler(c *gin.Context) {
	articleRule := regexp.MustCompile(`^\/(\d+).html$`)
	params := articleRule.FindStringSubmatch(c.Request.URL.Path)
	if len(params) == 2 {
		articleHandler(c, params[1])
		return
	}
	c.JSON(404, gin.H{"error": "404"})
}
