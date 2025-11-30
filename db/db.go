package db

import (
	"context"
	"encoding/json"
	"fmt"
	"mealmate-agent/models"
	"os"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
)

type MilvusDatabase struct {
	Client   *client.Client
	Embedder *ark.Embedder
	Indexer  *milvus.Indexer
}

func NewMilvusDatabase(ctx context.Context, milvusClient *client.Client, embedder *ark.Embedder) *MilvusDatabase {
	indexer := NewEventIndexer(ctx, milvusClient, embedder)
	return &MilvusDatabase{
		Client:   milvusClient,
		Embedder: embedder,
		Indexer:  indexer,
	}
}

/**
* @description: Index events from database to milvus
* @param ctx context.Context
* @return nil if success, error if failed
 */
func (db *MilvusDatabase) IndexEventsFromDatabase(ctx context.Context, config models.SyncConfig) (int, error) {
	SupabaseApiUrl := os.Getenv("SUPABASE_API_URL")
	SupabaseApiKey := os.Getenv("SUPABASE_API_KEY")
	supabaseClient := NewSupabaseClient(SupabaseApiUrl, SupabaseApiKey)
	data, _, err := supabaseClient.From("event").Select("*", "", false).Filter("user_id", "eq", config.UserID).Execute()
	if err != nil {
		return 0, err
	}
	fmt.Println("result:", string(data))

	var events []models.Event
	err = json.Unmarshal(data, &events)
	if err != nil {
		return 0, err
	}
	if len(events) < 1 {
		return 0, fmt.Errorf("no event found")
	}

	if db.Indexer == nil {
		db.Indexer = NewEventIndexer(ctx, db.Client, db.Embedder)
	}
	err = db.SyncEventToMilvus(ctx, &events)
	if err != nil {
		return 0, err
	}
	return len(events), nil
}

func (db *MilvusDatabase) SyncEventToMilvus(ctx context.Context, events *[]models.Event) error {
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
	_, err := db.Indexer.Store(ctx, docs)
	hlog.SystemLogger().Debug("Indexed events to Milvus:", len(docs))
	return err
}
