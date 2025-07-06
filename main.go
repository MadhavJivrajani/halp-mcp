// Server implementation largely inspired from: https://github.com/modelcontextprotocol/go-sdk/blob/main/examples/hello/main.go

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const mcpNamespace = "halpmcp"

type HalpArgs struct {
	Message string `json:"message"`
}

func ExecuteHalp(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[HalpArgs]) (*mcp.CallToolResultFor[struct{}], error) {
	msg := params.Arguments.Message
	_, err := createAndDeleteConfigMap(ctx, msg, mcpNamespace)
	if err != nil {
		log.Fatal(err)
	}

	return &mcp.CallToolResultFor[struct{}]{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("Message %s has been spent, godspeed.", msg),
			},
		},
	}, nil
}

func PromptHalp(_ context.Context, ss *mcp.ServerSession, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Description: "Halp message prompt",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: "Send an SOS message with content " + params.Arguments["message"]}},
		},
	}, nil
}

func main() {
	server := mcp.NewServer("halp-mcp", "v0.0.1", nil)
	server.AddTools(mcp.NewServerTool("halp", "send halp message", ExecuteHalp, mcp.Input(
		mcp.Property("message", mcp.Description("the halp message to send")),
	)))
	server.AddPrompts(&mcp.ServerPrompt{
		Prompt:  &mcp.Prompt{Name: "halp"},
		Handler: PromptHalp,
	})

	t := mcp.NewLoggingTransport(mcp.NewStdioTransport(), os.Stderr)
	if err := server.Run(context.Background(), t); err != nil {
		log.Printf("Server failed: %v", err)
	}
}
