package event

import (
	"context"
	"net/http"

	"mealmate-agent/db"
	"mealmate-agent/models"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
)

func Register(h *server.Hertz, milvusDB *db.MilvusDatabase) {
	// Register event-related routes here in the future
	h.POST("/event", func(ctx context.Context, c *app.RequestContext) {
		EventPostHandler(ctx, c, milvusDB)
	})
	h.POST("/event/sync", func(ctx context.Context, c *app.RequestContext) {
		EventSyncHandler(ctx, c, milvusDB)
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

	count, err := milvusDB.IndexEventsFromDatabase(ctx, config)
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
