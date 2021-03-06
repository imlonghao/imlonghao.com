package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gomarkdown/markdown"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gorilla/feeds"
	"github.com/tdewolff/minify/v2"
	minifycss "github.com/tdewolff/minify/v2/css"
	minifyhtml "github.com/tdewolff/minify/v2/html"
	minifyjs "github.com/tdewolff/minify/v2/js"
)

type articleModel struct {
	ID             int
	Title          string
	Description    string
	Content        string
	CreatedTime    time.Time
	CreatedTimeStr string
}
type linkModel struct {
	Name string
	Link string
}

var (
	articles []articleModel
	links    []linkModel
	tmpl     *template.Template
	ver      string
	wg       sync.WaitGroup
	lazyLoad *regexp.Regexp
)

const letters = "abcdefghijklmnopqrstuvwxyz0123456789"

func e(ee error) {
	if ee != nil {
		log.Fatal(ee)
	}
}

func atoi(s string) int {
	i, err := strconv.Atoi(s)
	e(err)
	return i
}

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func readArticle(filename string) (string, string, time.Time, string) {
	content, err := ioutil.ReadFile("posts/" + filename)
	e(err)

	line := strings.Split(string(content), "\n")
	title := line[1]
	desc := line[2]
	t := time.Unix(int64(atoi(line[3])), 0)

	content = []byte(strings.Join(line[5:], "\n"))

	htmlFlags := mdhtml.CommonFlags | mdhtml.NofollowLinks
	opts := mdhtml.RendererOptions{Flags: htmlFlags}
	renderer := mdhtml.NewRenderer(opts)
	content = markdown.ToHTML(content, nil, renderer)

	content = lazyLoad.ReplaceAll(content, []byte(`<img class="lazy" data-src="$1">`))

	return title, desc, t, string(content)
}

func loadArticles() {
	files, err := ioutil.ReadDir("posts")
	e(err)
	for _, file := range files {
		id, err := strconv.Atoi(strings.TrimSuffix(file.Name(), ".md"))
		e(err)
		title, desc, t, content := readArticle(file.Name())
		articles = append(articles, articleModel{
			id,
			title,
			desc,
			content,
			t,
			t.Format("2006-01-02 15:04:05"),
		})
	}
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].ID > articles[j].ID
	})
}

func loadLinks() {
	linksFile, err := ioutil.ReadFile("links.json")
	e(err)
	var linksString [][]string
	json.Unmarshal(linksFile, &linksString)
	for _, link := range linksString {
		links = append(links, linkModel{
			link[0],
			link[1],
		})
	}
}

func staticHandler(m *minify.M, ext string, filenames ...string) {
	var f []byte
	for _, filename := range filenames {
		file, err := ioutil.ReadFile(fmt.Sprintf("static/%s.%s", filename, ext))
		e(err)
		f = append(f, file...)
	}
	f, err := m.Bytes(ext, f)
	e(err)
	err = ioutil.WriteFile(fmt.Sprintf("dist/static/%s.%s", ver, ext), f, 0644)
	e(err)
}

func indexGenerator() {
	defer wg.Done()
	f, err := os.OpenFile("dist/index.html", os.O_RDWR|os.O_CREATE, 0644)
	e(err)
	m := minify.New()
	m.AddFunc("text/html", minifyhtml.Minify)
	mw := m.Writer("text/html", f)
	err = tmpl.ExecuteTemplate(mw, "index.html", map[string]interface{}{
		"title":       "首页",
		"description": "非专业人士的非专业博客",
		"articles":    articles,
		"links":       links,
		"ver":         ver,
	})
	e(err)
	mw.Close()
	f.Close()
}

func articleGenerator() {
	defer wg.Done()
	for _, article := range articles {
		f, err := os.OpenFile(fmt.Sprintf("dist/%d.html", article.ID), os.O_RDWR|os.O_CREATE, 0644)
		e(err)
		m := minify.New()
		m.AddFunc("text/html", minifyhtml.Minify)
		mw := m.Writer("text/html", f)
		err = tmpl.ExecuteTemplate(mw, "article.html", map[string]interface{}{
			"id":          article.ID,
			"title":       article.Title,
			"description": article.Description,
			"content":     template.HTML(article.Content),
			"ver":         ver,
		})
		e(err)
		mw.Close()
		f.Close()
	}
}

func feedGenerator() {
	defer wg.Done()
	feed := feeds.Feed{
		Title:       "imlonghao",
		Link:        &feeds.Link{Href: "https://imlonghao.com/"},
		Description: "非专业人士的非专业博客",
		Created:     time.Now(),
	}
	for _, article := range articles {
		url := fmt.Sprintf("https://imlonghao.com/%d.html", article.ID)
		feed.Items = append(feed.Items, &feeds.Item{
			Id:          url,
			Title:       article.Title,
			Link:        &feeds.Link{Href: url},
			Description: article.Description,
			Created:     article.CreatedTime,
		})
	}
	rss, _ := feed.ToRss()
	ioutil.WriteFile("dist/feed.xml", []byte(rss), 0644)
}

func sitemapGenerator() {
	defer wg.Done()
	sitemap := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.sitemaps.org/schemas/sitemap/0.9 http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd" xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	<url>
		<loc>https://imlonghao.com/</loc>
		<changefreq>daily</changefreq>
		<priority>1.0</priority>
	</url>
`)
	for _, article := range articles {
		s := fmt.Sprintf(`    <url>
        <loc>https://imlonghao.com/%d.html</loc>
        <changefreq>monthly</changefreq>
        <priority>0.5</priority>
	</url>
`, article.ID)
		sitemap += s
	}
	sitemap += `</urlset>`
	ioutil.WriteFile("dist/sitemap.xml", []byte(sitemap), 0644)
}

func staticFileGenerator() {
	defer wg.Done()
	m := minify.New()
	m.AddFunc("css", minifycss.Minify)
	m.AddFunc("js", minifyjs.Minify)
	staticHandler(m, "js", "prism", "disqus", "global")
	staticHandler(m, "css", "style", "prism")
}

func init() {
	rand.Seed(time.Now().UnixNano())
	tmpl, _ = template.ParseGlob("views/*.html")
	ver = randString(6)
	wg.Add(5)
	lazyLoad = regexp.MustCompile(`<img src="(.*?)" alt="" />`)
}

func main() {
	loadArticles()
	loadLinks()
	go indexGenerator()
	go articleGenerator()
	go feedGenerator()
	go sitemapGenerator()
	go staticFileGenerator()
	wg.Wait()
}
