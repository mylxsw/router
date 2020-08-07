package main

import (
	"log"
	"net/http"

	"github.com/mylxsw/container"
	"github.com/mylxsw/router/web"
)

func main() {
	mw := web.NewRequestMiddleware()

	cc := container.New()
	router := web.NewRouter(cc, web.DefaultConfig(), mw.AccessLog())

	router.Get("/", func(ctx web.Context) web.Response {
		return ctx.API("000000", "ok", web.M{
			"id": 12445,
		})
	})

	router.Group("/api", func(router *web.Router) {
		router.Group("/books", func(router *web.Router) {
			router.Get("/{id}", func(ctx web.Context) web.Response {
				return ctx.API("000000", "ok", web.M{
					"book_id": ctx.PathVar("id"),
				})
			})
			router.Put("{id}", func(ctx web.Context) web.Response {
				return ctx.API("000000", "ok", web.M{
					"book_id": ctx.PathVar("id"),
					"updated": true,
				})
			})
			router.Post("/", func(ctx web.Context, req web.Request) web.Response {
				return ctx.API("000000", "ok", web.M{
					"title":   req.Input("title"),
					"content": req.Input("content"),
				}).WithCode(http.StatusCreated)
			}).WithDecorators(mw.BeforeInterceptor(func(ctx web.Context) web.Response {
				log.Println("before-3")
				return nil
			}))
		}, mw.BeforeInterceptor(func(ctx web.Context) web.Response {
			log.Println("before-1")
			return nil
		}), mw.BeforeInterceptor(func(ctx web.Context) web.Response {
			log.Println("before-2")
			return nil
		}))
	}, mw.BeforeInterceptor(func(ctx web.Context) web.Response {
		log.Println("before-0")
		return nil
	}))

	for _, r := range router.Routes() {
		log.Printf("route: %s", r.String())
	}

	if err := http.ListenAndServe(":9999", router); err != nil {
		panic(err)
	}
}
