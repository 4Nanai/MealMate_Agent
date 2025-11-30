package pipeline

import (
	"context"

	"github.com/cloudwego/eino-ext/components/embedding/ark"
	"github.com/cloudwego/eino/compose"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
)

type EventAgentState struct {
	History map[string]any
}

/**
* @description: Build the MealMateAgent
* @param ctx context.Context
* @return r compose.Runnable[string, string], err error
* @return nil if success, error if failed
 */
func BuildMealMateAgent(ctx context.Context, embedder *ark.Embedder, milvusClient *client.Client) (r compose.Runnable[string, string], err error) {
	const (
		UserProfileRetriever = "UserProfileRetriever"
		UserProfileGen       = "UserProfileGen"
		EventChatTemplate    = "EventChatTemplate"
		ChatModel            = "ChatModel"
		outputFormatHandler  = "outputFormatHandler"
	)
	g := compose.NewGraph[string, string](compose.WithGenLocalState(func(ctx context.Context) (state EventAgentState) {
		return EventAgentState{
			History: map[string]any{},
		}
	}))

	// Create Event Retriever Node
	dynamicRetriever := NewDynamicFilterRetriever(embedder, milvusClient)
	_ = g.AddRetrieverNode(UserProfileRetriever, dynamicRetriever)
	_ = g.AddLambdaNode(UserProfileGen, compose.InvokableLambda(genUserProfile))
	eventChatTemplateKeyOfChatTemplate, err := newChatTemplate(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatTemplateNode(EventChatTemplate, eventChatTemplateKeyOfChatTemplate, compose.WithStatePreHandler(chatTemplatePreHandler))
	chatModelKeyOfChatModel, err := newChatModel(ctx)
	if err != nil {
		return nil, err
	}
	_ = g.AddChatModelNode(ChatModel, chatModelKeyOfChatModel)
	_ = g.AddLambdaNode(outputFormatHandler, compose.InvokableLambda(chatOutputHandler))
	_ = g.AddEdge(compose.START, UserProfileRetriever)
	_ = g.AddEdge(outputFormatHandler, compose.END)
	_ = g.AddEdge(UserProfileRetriever, UserProfileGen)
	_ = g.AddEdge(UserProfileGen, EventChatTemplate)
	_ = g.AddEdge(EventChatTemplate, ChatModel)
	_ = g.AddEdge(ChatModel, outputFormatHandler)
	r, err = g.Compile(ctx, compose.WithGraphName("MealMateAgent"), compose.WithNodeTriggerMode(compose.AnyPredecessor))
	if err != nil {
		return nil, err
	}
	return r, err
}
