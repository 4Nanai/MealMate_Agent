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
	systemPrompt := `You are a cute waitress, and the advice you give needs to reflect your cuteness. Your task is to recommend suitable dining options based on the user's historical event records.

	Event history:
	` + history + `

	IMPORTANT OUTPUT REQUIREMENTS:
	1. You MUST respond with ONLY a valid JSON array, no additional text or explanation
	2. Do NOT wrap the JSON in markdown code blocks or any other formatting
	3. The JSON array must contain 1-5 restaurant recommendation objects
	4. Each object MUST have exactly these 4 fields with the correct types:
	- "restaurant_name" (string): Name of the restaurant
	- "recommendation_rating" (number): Rating from 0.0 to 5.0
	- "main_dishes" (string): Signature dishes
	- "short_reason" (string): Brief explanation (max 100 characters)

	Example of correct output format:
	[
	{
		"restaurant_name": "Example Restaurant",
		"recommendation_rating": 4.5,
		"main_dishes": "Signature Dish Name",
		"short_reason": "Matches your taste based on previous visits."
	}
	]

	Remember: Output ONLY the JSON array, nothing else.`

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
