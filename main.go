package main

import (
	"context"
	"encoding/json"
	"fmt"
	"mealmate-agent/models"
	"mealmate-agent/pipeline"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	ctx := context.Background()

	// Initialize shared embedder and milvus client
	InitEmbedder(ctx)
	InitMilvusClient(ctx)

	// Pass shared instances to pipeline package
	pipeline.SetSharedInstances(GlobalEmbedder, MilvusCli)

	runnable, err := pipeline.BuildMealMateAgent(ctx)
	if err != nil {
		panic(err)
	}
	output, err := runnable.Invoke(ctx, `{
	"user_id": "xs90",
	"username": "Hello World",
	"user_prompt": "I wanna some fried chicken."}`)
	if err != nil {
		panic(err)
	}
	fmt.Println("output:", output)
}

/**
* @description: Index events from database to milvus
* @param ctx context.Context
* @return nil if success, error if failed
 */
func IndexEventsFromDatabase(ctx context.Context) error {
	SupabaseApiUrl := os.Getenv("SUPABASE_API_URL")
	SupabaseApiKey := os.Getenv("SUPABASE_API_KEY")
	supabaseClient := NewSupabaseClient(SupabaseApiUrl, SupabaseApiKey)
	data, _, err := supabaseClient.From("event").Select("*", "", false).Filter("user_id", "eq", "xs90").Execute()
	if err != nil {
		return err
	}
	fmt.Println("result:", string(data))

	var events []models.Event
	err = json.Unmarshal(data, &events)
	if err != nil {
		return err
	}
	if len(events) < 1 {
		return fmt.Errorf("no event found")
	}

	// Use shared embedder and milvus client
	if MilvusCli == nil {
		InitMilvusClient(ctx)
	}
	if GlobalEmbedder == nil {
		InitEmbedder(ctx)
	}

	indexer := NewEventIndexer(ctx)
	err = syncEventToMilvus(ctx, &events, indexer)
	if err != nil {
		return err
	}
	return nil
}
