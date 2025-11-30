package main

import (
	"context"
	"mealmate-agent/biz/router"
	"mealmate-agent/db"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	ctx := context.Background()

	// Initialize Milvus client and embedder
	milvusClient := InitMilvusClient(ctx)
	hlog.SystemLogger().Info("Milvus client initialized")
	embedder, err := InitEmbedder(ctx)
	hlog.SystemLogger().Info("Embedder initialized")
	if err != nil {
		panic(err)
	}
	// Initialize MilvusDatabase
	milvusDB := db.NewMilvusDatabase(ctx, &milvusClient, embedder)
	hlog.SystemLogger().Info("MilvusDatabase initialized")

	// Start automatic sync task
	milvusDB.StartAutoSync(ctx)
	hlog.SystemLogger().Info("Automatic sync task started")

	// Start Hertz server
	h := server.Default(server.WithHostPorts("127.0.0.1:8080"))

	router.RegisterRoutes(h, milvusDB)

	h.Spin()
}
