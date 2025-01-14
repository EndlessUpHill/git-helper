package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// Add this interface
type openAIClient interface {
	CreateChatCompletion(context.Context, openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

type CommitGenerator struct {
	client openAIClient
}

func NewCommitGenerator(apiKey string) *CommitGenerator {
	return &CommitGenerator{
		client: openai.NewClient(apiKey),
	}
}

func (g *CommitGenerator) GenerateCommitMessage(diff string) (string, error) {
	prompt := fmt.Sprintf(`Generate a conventional commit message for the following git diff:

%s

The commit message should:
1. Follow the format: <type>(<optional scope>): <description>
2. Use one of these types: feat, fix, docs, style, refactor, test, chore
3. Be concise but descriptive
4. Focus on the "what" and "why" rather than the "how"
5. Use imperative mood ("add" not "added")

Return only the commit message without any additional text.`, diff)

	resp, err := g.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.7,
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate commit message: %w", err)
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
} 