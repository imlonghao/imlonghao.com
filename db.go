package main

import (
	"fmt"
	"os"

	"github.com/jackc/pgx"
)

var pool *pgx.ConnPool

func afterConnect(conn *pgx.Conn) (err error) {
	_, _ = conn.Prepare("getArticles", "SELECT id,title FROM articles WHERE time<now() ORDER BY id DESC")
	_, _ = conn.Prepare("getLinks", "SELECT link,name FROM links ORDER BY random()")
	_, _ = conn.Prepare("getArticle", "SELECT id, title, description, content FROM articles WHERE id=$1 and time<now()")
	_, _ = conn.Prepare("getSitemap", `SELECT id, to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS+08:00') FROM articles WHERE time<now() ORDER BY id DESC`)
	_, _ = conn.Prepare("getFeed", "SELECT id, title, description, time FROM articles WHERE time<now() ORDER BY id DESC LIMIT 10")
	return
}

func dbInit() {
	connPoolConfig := pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     os.Getenv("DBHOST"),
			User:     os.Getenv("DBUSER"),
			Password: os.Getenv("DBPASS"),
			Database: os.Getenv("DBNAME"),
		},
		MaxConnections: 10,
		AfterConnect:   afterConnect,
	}
	var err error
	pool, err = pgx.NewConnPool(connPoolConfig)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
