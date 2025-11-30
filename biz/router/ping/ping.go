package ping

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func Register(h *server.Hertz) {
	h.GET("/ping", PingHandler)
}

func PingHandler(ctx context.Context, c *app.RequestContext) {
	hlog.SystemLogger().Info("ping received")
	c.JSON(consts.StatusOK, utils.H{"message": "pong"})
}
