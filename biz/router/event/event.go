package event

import (
	"context"
	"net/http"

	"mealmate-agent/db"
	"mealmate-agent/models"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
)

func Register(h *server.Hertz, milvusDB *db.MilvusDatabase, runnable *compose.Runnable[string, string]) {
	// Register event-related routes here in the future
	h.POST("/events", func(ctx context.Context, c *app.RequestContext) {
		EventPostHandler(ctx, c, milvusDB)
	})
	h.POST("/events/sync", func(ctx context.Context, c *app.RequestContext) {
		EventSyncHandler(ctx, c, milvusDB)
	})
	h.POST("/events/ai", func(ctx context.Context, c *app.RequestContext) {
		CallEventAgent(ctx, c, runnable)
	})
}

func EventPostHandler(ctx context.Context, c *app.RequestContext, milvusDB *db.MilvusDatabase) {
	var event models.Event
	var err error

	// Validate and bind the request body to the Event struct
	if err = c.BindAndValidate(&event); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"error":  "Invalid request body",
			"detail": err.Error(),
		})
		return
	}

	hlog.SystemLogger().Info("Event received:", event)

	err = milvusDB.SyncEventToMilvus(ctx, &[]models.Event{event})
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"error":  "Failed to index event",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, utils.H{
		"message": "Event received successfully",
	})

	hlog.SystemLogger().Info("Event indexed successfully:", event)
}

func EventSyncHandler(ctx context.Context, c *app.RequestContext, milvusDB *db.MilvusDatabase) {
	var err error

	var config models.SyncConfig

	// Validate and bind the request body to the SyncConfig struct
	if err = c.BindAndValidate(&config); err != nil {
		c.JSON(http.StatusBadRequest, utils.H{
			"error":  "Invalid request body",
			"detail": err.Error(),
		})
		return
	}

	hlog.SystemLogger().Info("Starting event sync for user:", config.UserID)

	count, err := milvusDB.ManuallySyncDatabase(ctx, config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"error":  "Failed to sync events",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, utils.H{
		"message": "Events synced successfully",
		"count":   count,
	})

	hlog.SystemLogger().Info("Event sync completed for user:", config.UserID, "Count:", count)
}

func CallEventAgent(ctx context.Context, c *app.RequestContext, runnable *compose.Runnable[string, string]) {
	// Get raw request body
	body := c.Request.Body()

	// Validate body length
	if len(body) == 0 {
		c.JSON(http.StatusBadRequest, utils.H{
			"error": "Request body cannot be empty",
		})
		return
	}

	output, err := (*runnable).Invoke(ctx, string(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.H{
			"error":  "Failed to process request",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, utils.H{
		"response": output,
	})
}
