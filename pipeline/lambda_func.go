package pipeline

import (
	"context"

	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/hertz/pkg/common/hlog"
)

// genUserProfile component initialization function of node 'UserProfileGen' in graph 'MealMateAgent'
func genUserProfile(ctx context.Context, input []*schema.Document) (output map[string]any, err error) {
	if len(input) == 0 {
		return nil, nil
	}
	output = make(map[string]any)
	var history string
	for _, doc := range input {
		history += doc.Content + "\n"
	}
	output["history"] = history
	return output, nil
}

// chatOutputHandler component initialization function of node 'outputFormatHandler' in graph 'MealMateAgent'
func chatOutputHandler(ctx context.Context, input *schema.Message) (output string, err error) {
	content := input.Content
	hlog.SystemLogger().Info("AI Response:", content)
	return content, nil
}
