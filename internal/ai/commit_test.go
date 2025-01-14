package ai

import (
	"context"
	"strings"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock OpenAI client
type mockOpenAIClient struct {
	mock.Mock
}

func (m *mockOpenAIClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(openai.ChatCompletionResponse), args.Error(1)
}

func TestGenerateCommitMessage(t *testing.T) {
	tests := []struct {
		name        string
		diff        string
		mockResp    string
		expectError bool
	}{
		{
			name: "successful commit message generation",
			diff: `diff --git a/cmd/root.go b/cmd/root.go
+       fmt.Printf("OpenAI API key present: %v\n", viper.GetString("openai_api_key") != "")`,
			mockResp:    "feat(config): add OpenAI API key validation",
			expectError: false,
		},
		{
			name:        "empty diff",
			diff:        "",
			mockResp:    "chore: no changes detected",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockOpenAIClient{}
			generator := &CommitGenerator{client: mockClient}

			// Setup mock response
			mockClient.On("CreateChatCompletion", mock.Anything, mock.MatchedBy(func(req openai.ChatCompletionRequest) bool {
				return strings.Contains(req.Messages[0].Content, tt.diff)
			})).Return(openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: tt.mockResp,
						},
					},
				},
			}, nil)

			// Call the function
			msg, err := generator.GenerateCommitMessage(tt.diff)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.mockResp, msg)
			}

			mockClient.AssertExpectations(t)
		})
	}
} 