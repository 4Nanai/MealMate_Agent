package pipeline

import (
	"context"
	"os"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino/components/model"
)

// newChatModel component initialization function of node 'ChatModel' in graph 'MealMateAgent'
func newChatModel(ctx context.Context) (cm model.ChatModel, err error) {
	config := &ark.ChatModelConfig{
		APIKey: os.Getenv("ARK_API_KEY"),
		Model: os.Getenv("ARK_CHAT_MODEL"),
	}
	cm, err = ark.NewChatModel(ctx, config)
	if err != nil {
		return nil, err
	}
	return cm, nil
}
