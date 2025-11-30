package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino-ext/components/retriever/milvus"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

func newRetriever(ctx context.Context, embedder *ark.Embedder, milvusClient *client.Client) (*milvus.Retriever, error) {
	// check if shared embedder and milvus client are initialized
	if embedder == nil {
		panic("embedder not initialized, call SetSharedInstances first")
	}
	if milvusClient == nil {
		panic("milvus client not initialized, call SetSharedInstances first")
	}

	searchParam, err := entity.NewIndexHNSWSearchParam(10)
	if err != nil {
		panic(err)
	}
	r, err := milvus.NewRetriever(ctx, &milvus.RetrieverConfig{
		Client:      *milvusClient,
		Collection:  os.Getenv("MILVUS_EVENT_COLLECTION"),
		VectorField: "vector",
		OutputFields: []string{
			"event_id",
			"content",
			"meta_data",
			"user_id",
		},
		TopK:      3,
		Embedding: embedder,
		DocumentConverter: func(ctx context.Context, doc client.SearchResult) ([]*schema.Document, error) {
			var err error
			result := make([]*schema.Document, doc.IDs.Len())
			for i := range result {
				result[i] = &schema.Document{
					MetaData: make(map[string]any),
				}
			}
			for _, field := range doc.Fields {
				switch field.Name() {
				case "event_id":
					for i, document := range result {
						document.ID, err = doc.IDs.GetAsString(i)
						if err != nil {
							return nil, fmt.Errorf("failed to get id: %w", err)
						}
					}
				case "content":
					for i, document := range result {
						document.Content, err = field.GetAsString(i)
						if err != nil {
							return nil, fmt.Errorf("failed to get content: %w", err)
						}
					}
				case "meta_data":
					for i, document := range result {
						b, err := field.Get(i)
						bytes, ok := b.([]byte)
						if !ok {
							return nil, fmt.Errorf("failed to get metadata: %w", err)
						}
						if err := sonic.Unmarshal(bytes, &document.MetaData); err != nil {
							return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
						}
					}
				case "user_id":
					for i, document := range result {
						document.MetaData["user_id"], err = field.GetAsString(i)
						if err != nil {
							return nil, fmt.Errorf("failed to get user_id: %w", err)
						}
					}
				}
			}
			return result, nil
		},
		VectorConverter: func(ctx context.Context, vectors [][]float64) ([]entity.Vector, error) {
			vecs := make([]entity.Vector, len(vectors))
			for i, vector := range vectors {
				float32Vec := make([]float32, len(vector))
				for j, v := range vector {
					float32Vec[j] = float32(v)
				}
				vecs[i] = entity.FloatVector(float32Vec)
			}
			return vecs, nil
		},
		MetricType: entity.COSINE,
		Sp:         searchParam,
	})
	if err != nil {
		return nil, err
	}
	return r, nil
}

func WithFilter(filterExpr string) retriever.Option {
	return retriever.WrapImplSpecificOptFn(func(io *milvus.ImplOptions) {
		io.Filter = filterExpr
	})
}

// Input JSON for retriever
type RetrieverInput struct {
	UserPrompt string `json:"user_prompt"`
	UserID     string `json:"user_id"`
	Username   string `json:"username"`
}

// Wrapped retriever to support dynamic filter
type DynamicFilterRetriever struct {
	baseRetriever retriever.Retriever
}

func NewDynamicFilterRetriever(embedder *ark.Embedder, milvusClient *client.Client) *DynamicFilterRetriever {
	base, err := newRetriever(context.Background(), embedder, milvusClient)
	if err != nil {
		panic(err)
	}
	return &DynamicFilterRetriever{
		baseRetriever: base,
	}
}

// Implement the retriever.Retriever interface
func (r *DynamicFilterRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	var input RetrieverInput
	if err := json.Unmarshal([]byte(query), &input); err == nil {
		actualQuery := input.UserPrompt
		if actualQuery == "" {
			return nil, fmt.Errorf("user prompt is empty")
		}
		if input.UserID == "" {
			return nil, fmt.Errorf("user id is empty")
		}
		if input.Username == "" {
			return nil, fmt.Errorf("username is empty")
		}
		compose.ProcessState(ctx, func(ctx context.Context, state EventAgentState) error {
			state.History["user_id"] = input.UserID
			state.History["user_prompt"] = input.UserPrompt
			state.History["username"] = input.Username
			return nil
		})
		filterExpr := fmt.Sprintf("user_id == \"%s\"", input.UserID)
		opts = append(opts, WithFilter(filterExpr))

		return r.baseRetriever.Retrieve(ctx, actualQuery, opts...)
	}
	return nil, fmt.Errorf("input is not a valid json")
}
