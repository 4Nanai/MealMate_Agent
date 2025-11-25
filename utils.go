package main

import (
	"context"
	"fmt"
	"mealmate-agent/models"

	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino/schema"
)

func syncEventToMilvus(ctx context.Context, events *[]models.Event, indexer *milvus.Indexer) error {
	docs := make([]*schema.Document, 0)
	for _, event := range *events {
		eventId := fmt.Sprintf("%d", event.ID)
		text := fmt.Sprintf("%s %s", event.RestaurantName, event.Message)
		doc := &schema.Document{
			ID:      eventId,
			Content: text,
			MetaData: map[string]any{
				"user_id":   event.UserID,
				"latitude":  event.RestaurantCoordinates.Latitude,
				"longitude": event.RestaurantCoordinates.Longitude,
				"create_at": event.CreatedAt,
				"schedule":  event.ScheduleTime,
			},
		}
		docs = append(docs, doc)
	}
	if len(docs) == 0 {
		return nil
	}
	_, err := indexer.Store(ctx, docs)
	return err
}

