package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mylxsw/container"
	"github.com/mylxsw/router/web"
)

type Logger struct{}

func (Logger) Debugf(format string, v ...interface{}) {
	log.Printf(fmt.Sprintf("[DEBUG] %s", format), v...)
}

func (Logger) Errorf(format string, v ...interface{}) {
	log.Printf(fmt.Sprintf("[ERROR] %s", format), v...)
}

func main() {
	logger := Logger{}
	mw := web.NewRequestMiddleware()

	cc := container.New()
	router := web.NewRouter(cc, web.DefaultConfig(), mw.AccessLog(logger)).
		WithLogger(logger).
		WithRouteNotFoundHandler(func(wtx web.Context, route web.RealRoute) web.Response {
			return wtx.JSONWithCode(web.M{"error": "no such route"}, http.StatusNotFound)
		}).
		WithExceptionHandler(func(wtx web.Context, err error) web.Response {
			logger.Errorf("request exception occurs: %v", err)
			return wtx.JSONWithCode(web.M{"error": fmt.Sprintf("Error: %v", err)}, http.StatusInternalServerError)
		})

	router.Get("/", func(wtx web.Context) web.Response {
		return wtx.API("000000", "ok", web.M{
			"id": 12445,
		})
	})

	router.Get("/errors", func(wtx web.Context) web.Response {
		panic("oops")
	})

	router.Group("/api", func(router *web.Router) {
		router.Group("/books", func(router *web.Router) {
			router.Get("/{id}", func(wtx web.Context) web.Response {
				return wtx.API("000000", "ok", web.M{
					"book_id": wtx.PathVar("id"),
				})
			})
			router.Put("{id}", func(wtx web.Context) web.Response {
				return wtx.API("000000", "ok", web.M{
					"book_id": wtx.PathVar("id"),
					"updated": true,
				})
			})
			router.Post("/", func(wtx web.Context, req web.Request) web.Response {
				return wtx.API("000000", "ok", web.M{
					"title":   req.Input("title"),
					"content": req.Input("content"),
				}).WithCode(http.StatusCreated)
			}).WithDecorators(mw.BeforeInterceptor(func(wtx web.Context) web.Response {
				logger.Debugf("before-3")
				return nil
			}))
		}, mw.BeforeInterceptor(func(wtx web.Context) web.Response {
			logger.Debugf("before-1")
			return nil
		}), mw.BeforeInterceptor(func(wtx web.Context) web.Response {
			logger.Debugf("before-2")
			return nil
		}))
	}, mw.BeforeInterceptor(func(wtx web.Context) web.Response {
		logger.Debugf("before-0")
		return nil
	}))

	for _, r := range router.Routes() {
		logger.Debugf("route: %s", r.String())
	}

	if err := router.ListenAndServe(":9999"); err != nil {
		panic(err)
	}
}
