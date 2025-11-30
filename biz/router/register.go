package router

import (
	"mealmate-agent/biz/router/event"
	"mealmate-agent/biz/router/ping"
	"mealmate-agent/db"

	"github.com/cloudwego/hertz/pkg/app/server"
)

func RegisterRoutes(h *server.Hertz, milvusDB *db.MilvusDatabase) {
	ping.Register(h)
	event.Register(h, milvusDB)
}
