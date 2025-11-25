package main

import (
	"context"
	"os"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
)

// GlobalEmbedder is the global shared embedder instance
var GlobalEmbedder *ark.Embedder

/**
* @description: Initialize the embedder
* @param ctx context.Context
* @return nil if success, error if failed
 */
func InitEmbedder(ctx context.Context) {
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: os.Getenv("ARK_API_KEY"),
		Model:  os.Getenv("ARK_EMBEDDER_MODEL"),
	})
	if err != nil {
		panic(err)
	}
	GlobalEmbedder = embedder
}

/**
* @description: Get the global shared embedder instance
* @return the global shared embedder instance
 */
func GetEmbedder() *ark.Embedder {
	return GlobalEmbedder
}
