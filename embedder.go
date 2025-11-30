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
* @return embedder instance and error
 */
func InitEmbedder(ctx context.Context) (*ark.Embedder, error) {
	embedder, err := ark.NewEmbedder(ctx, &ark.EmbeddingConfig{
		APIKey: os.Getenv("ARK_API_KEY"),
		Model:  os.Getenv("ARK_EMBEDDER_MODEL"),
	})
	if err != nil {
		return nil, err
	}
	GlobalEmbedder = embedder
	return embedder, nil
}

/**
* @description: Get the global shared embedder instance
* @return the global shared embedder instance
 */
func GetEmbedder() *ark.Embedder {
	return GlobalEmbedder
}
