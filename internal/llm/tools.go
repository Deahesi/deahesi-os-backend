package llm

import "github.com/OpenRouterTeam/go-sdk/models/components"

// import (
// 	openrouter "github.com/OpenRouterTeam/go-sdk"
// 	"github.com/OpenRouterTeam/go-sdk/models/components"
// )

// var readFileTool = components.CreateChatFunctionToolChatFunctionToolFunction(
// 	components.ChatFunctionToolFunction{
// 		Type: components.ChatFunctionToolType("function"),

// 		Function: components.ChatFunctionToolFunctionFunction{
// 			Name:        "read_file",
// 			Description: openrouter.Pointer("Read contents of a file"),

// 			Parameters: map[string]any{
// 				"type": "object",

// 				"properties": map[string]any{
// 					"path": map[string]any{
// 						"type": "string",
// 					},
// 				},

// 				"required": []string{
// 					"path",
// 				},
// 			},
// 		},
// 	},
// )

var tools = []components.ChatFunctionTool{
	// readFileTool,
}
