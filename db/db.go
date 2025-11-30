package db

import (
	"context"
	"encoding/json"
	"fmt"
	"mealmate-agent/models"
	"os"
	"time"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/supabase-community/supabase-go"
)

type MilvusDatabase struct {
	Client   *client.Client
	Embedder *ark.Embedder
	Indexer  *milvus.Indexer
	Supabase *supabase.Client
}

func NewMilvusDatabase(ctx context.Context, milvusClient *client.Client, embedder *ark.Embedder) *MilvusDatabase {
	indexer := NewEventIndexer(ctx, milvusClient, embedder)
	SupabaseApiUrl := os.Getenv("SUPABASE_API_URL")
	SupabaseApiKey := os.Getenv("SUPABASE_API_KEY")
	supabaseClient := NewSupabaseClient(SupabaseApiUrl, SupabaseApiKey)
	return &MilvusDatabase{
		Client:   milvusClient,
		Embedder: embedder,
		Indexer:  indexer,
		Supabase: supabaseClient,
	}
}

/**
* @description: Index events from database to milvus
* @param ctx context.Context
* @return nil if success, error if failed
 */
func (db *MilvusDatabase) ManuallySyncDatabase(ctx context.Context, config models.SyncConfig) (int, error) {
	data, _, err := db.Supabase.From("event").Select("*", "", false).Filter("user_id", "eq", config.UserID).Execute()
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

func (db *MilvusDatabase) AutomaticSyncDatabase(ctx context.Context) error {
	// Calculate the time one minute ago
	oneMinuteAgo := time.Now().UTC().Add(-1 * time.Minute).Format(time.RFC3339)
	hlog.SystemLogger().Infof("Fetching events created after: %s (UTC)", oneMinuteAgo)

	// Fetch events created within the last minute from Supabase
	data, _, err := db.Supabase.From("event").Select("*", "", false).
		Filter("created_at", "gte", oneMinuteAgo).
		Execute()
	if err != nil {
		hlog.SystemLogger().Errorf("Failed to fetch events from Supabase: %v", err)
		return err
	}

	var events []models.Event
	err = json.Unmarshal(data, &events)
	if err != nil {
		hlog.SystemLogger().Errorf("Failed to unmarshal events: %v", err)
		return err
	}

	if len(events) == 0 {
		hlog.SystemLogger().Info("No new events to sync")
		return nil
	}

	hlog.SystemLogger().Infof("Found %d new events to sync", len(events))

	// Sync events to Milvus
	err = db.SyncEventToMilvus(ctx, &events)
	if err != nil {
		hlog.SystemLogger().Errorf("Failed to sync events to Milvus: %v", err)
		return err
	}

	hlog.SystemLogger().Infof("Successfully synced %d events to Milvus", len(events))
	return nil
}

// StartAutoSync starts a background goroutine that automatically syncs the database every minute.
func (db *MilvusDatabase) StartAutoSync(ctx context.Context) {
	hlog.SystemLogger().Info("Starting automatic sync task...")

	// Create a ticker that triggers every minute for testing purposes
	ticker := time.NewTicker(1 * time.Minute)

	// Run the scheduled task in the background
	go func() {
		defer ticker.Stop()

		// Run sync immediately
		if err := db.AutomaticSyncDatabase(ctx); err != nil {
			hlog.SystemLogger().Errorf("Initial sync failed: %v", err)
		} else {
			hlog.SystemLogger().Info("Initial sync completed successfully")
		}

		// Run sync periodically
		for {
			select {
			case <-ctx.Done():
				hlog.SystemLogger().Info("Auto sync task stopped due to context cancellation")
				return
			case <-ticker.C:
				hlog.SystemLogger().Info("Running scheduled sync...")
				if err := db.AutomaticSyncDatabase(ctx); err != nil {
					hlog.SystemLogger().Errorf("Scheduled sync failed: %v", err)
				} else {
					hlog.SystemLogger().Info("Scheduled sync completed successfully")
				}
			}
		}
	}()

	hlog.SystemLogger().Info("Automatic sync task started, will run every minute")
}
