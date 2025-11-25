package pipeline

import (
	"context"
	"log"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

type ChatTemplateImpl struct {
	config *ChatTemplateConfig
}

type ChatTemplateConfig struct {
}

// newChatTemplate component initialization function of node 'EventChatTemplate' in graph 'MealMateAgent'
func newChatTemplate(ctx context.Context) (ctp prompt.ChatTemplate, err error) {
	// TODO Modify component configuration here.
	config := &ChatTemplateConfig{}
	ctp = &ChatTemplateImpl{config: config}
	return ctp, nil
}

func (impl *ChatTemplateImpl) Format(ctx context.Context, vs map[string]any, opts ...prompt.Option) ([]*schema.Message, error) {
	history := vs["history"].(string)
	username := vs["username"].(string)
	userPrompt := vs["user_prompt"].(string)
	systemPrompt := "You are an intelligent dining recommendation assistant. Your task is to recommend suitable dining options based on the user's historical event records. Please ensure the recommendations match the user's taste and preferences.\nEvent history:\n" + history + "\n Your answer should be a JSON string, containing a list of objects, with each object containing \"restaurant_name\", \"recommendation_rating\", \"main_dishes\", \"short_reason\"."
	query := "I'm " + username + ", " + userPrompt
	messages := []*schema.Message{
		{
			Role:    schema.System,
			Content: systemPrompt,
		},
		{
			Role:    schema.User,
			Content: query,
		},
	}
	return messages, nil
}

func chatTemplatePreHandler(ctx context.Context, in map[string]any, state EventAgentState) (map[string]any, error) {
	userPrompt := state.History["user_prompt"]
	username := state.History["username"]
	in["user_prompt"] = userPrompt
	in["username"] = username
	log.Println("Updated input in chatTemplatePreHandler:", in)
	return in, nil
}
