package main

import (
	"context"
	"errors"
	"fmt"
	"go.deanishe.net/favicon"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	_ "github.com/mat/besticon/ico"
)

func main() {
	r := gin.New()
	r.Use(gin.Recovery())

	setupRouter(r)

	start(&http.Server{
		Addr:    fmt.Sprintf(":%s", env("FC_SERVER_PORT", "9000")),
		Handler: r,
	})
}

func setupRouter(r *gin.Engine) {
	r.GET("/ping", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("pong"))
	})
	r.GET("/fetch", func(c *gin.Context) {
		url, has := c.GetQuery("url")
		if !has {
			c.JSON(http.StatusOK, failed("missing required param 'url'"))
			return
		}
		icons, err := favicon.Find(url)
		if err != nil {
			c.JSON(http.StatusOK, failed(err.Error()))
			return
		}
		c.JSON(http.StatusOK, data(icons))
	})
	r.GET("/preview", func(c *gin.Context) {
		url, has := c.GetQuery("url")
		if !has {
			c.JSON(http.StatusOK, failed("missing required param 'url'"))
			return
		}
		icons, err := favicon.Find(url)
		if err != nil {
			c.JSON(http.StatusOK, failed(err.Error()))
			return
		}
		if len(icons) == 0 {
			c.JSON(http.StatusOK, failed("favicon not found"))
			return
		}
		// find largest icon
		var largest = icons[0]
		for _, icon := range icons {
			if icon.Width*icon.Height > largest.Width*largest.Height {
				largest = icon
			}
		}
		c.Redirect(http.StatusTemporaryRedirect, largest.URL)
	})
}

func start(srv *http.Server) {
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {

			log.Printf("listen: %s\n", err)
		}
	}()

	log.Printf("Start Server @ %s", srv.Addr)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Print("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server Shutdown:%s", err)
	}
	<-ctx.Done()
	log.Print("Server exiting")
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func failed(msg string) gin.H {
	return gin.H{
		"msg":       msg,
		"timestamp": time.Now().Unix(),
	}
}

func data(data interface{}) gin.H {
	return gin.H{
		"msg":       "success",
		"data":      data,
		"timestamp": time.Now().Unix(),
	}
}
