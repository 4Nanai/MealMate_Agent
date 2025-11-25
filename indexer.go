package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudwego/eino-ext/components/indexer/milvus"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

var eventFields = []*entity.Field{
	{
		Name:       "event_id",
		DataType:   entity.FieldTypeVarChar,
		PrimaryKey: true,
		AutoID:     false,
		TypeParams: map[string]string{
			"max_length": "256",
		},
	},
	{
		Name:     "vector",
		DataType: entity.FieldTypeFloatVector,
		TypeParams: map[string]string{
			"dim": "2560",
		},
	},
	{
		Name:     "content",
		DataType: entity.FieldTypeVarChar,
		TypeParams: map[string]string{
			"max_length":      "8192",
			"enable_analyzer": "false",
		},
	},
	{
		Name:     "meta_data",
		DataType: entity.FieldTypeJSON,
		TypeParams: map[string]string{
			"enable_analyzer": "false",
		},
	},
	{
		Name:     "user_id",
		DataType: entity.FieldTypeVarChar,
		TypeParams: map[string]string{
			"max_length":      "256",
			"enable_analyzer": "false",
		},
	},
}

func NewEventIndexer(ctx context.Context) *milvus.Indexer {
	indexer, err := milvus.NewIndexer(ctx, &milvus.IndexerConfig{
		Client:     MilvusCli,
		Collection: os.Getenv("MILVUS_EVENT_COLLECTION"),
		Embedding:  GlobalEmbedder,
		Fields:     eventFields,
		MetricType: milvus.COSINE,
		DocumentConverter: func(ctx context.Context, docs []*schema.Document, vectors [][]float64) ([]interface{}, error) {
			rows := make([]interface{}, 0, len(docs))
			for i, doc := range docs {
				userId := doc.MetaData["user_id"]
				if userId == nil {
					return nil, fmt.Errorf("user_id is missing in meta_data for document ID %s", doc.ID)
				}

				metaData := make(map[string]any)
				for k, v := range doc.MetaData {
					if k != "user_id" {
						metaData[k] = v
					}
				}

				// Convert []float64 to []float32 for Milvus
				vector32 := make([]float32, len(vectors[i]))
				for j, v := range vectors[i] {
					vector32[j] = float32(v)
				}

				row := map[string]interface{}{
					"event_id":  doc.ID,
					"vector":    vector32,
					"content":   doc.Content,
					"meta_data": metaData,
					"user_id":   userId,
				}
				rows = append(rows, row)
			}
			return rows, nil
		},
	})
	if err != nil {
		panic(err)
	}
	return indexer
}
