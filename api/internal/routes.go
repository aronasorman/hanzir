package internal

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"

	"github.com/samber/do"
)

func InitRoutes(i *do.Injector) *gin.Engine {
	r := gin.Default()

	// Enable CORS for Chrome extension
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/health", HealthCheckHandler(i))
	r.POST("/translate", TranslationHandler(i))

	return r
}

func HealthCheckHandler(i *do.Injector) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	}
}

type TranslationHandlerRequest struct {
	Characters string `json:"characters"`
}

type Definition struct {
	Type     string   `json:"type"`     // e.g., "verb", "noun", etc.
	Meanings []string `json:"meanings"` // list of meanings for this type
}

type TranslationHandlerResponse struct {
	Characters  string       `json:"characters"`
	Pinyin      string       `json:"pinyin"`
	Definitions []Definition `json:"definitions"`
}

var promptTemplate string = `For the Chinese characters "${characters}", provide:
1. The pinyin (with tone marks)
2. A detailed breakdown of definitions by part of speech (verb, noun, adjective, etc.)
Format the response as JSON only, no other text.
Example format:
{
    "pinyin": "nǐ hǎo",
    "definitions": [
        {
            "type": "greeting",
            "meanings": ["hello", "hi"]
        }
    ]
}`

func TranslationHandler(i *do.Injector) gin.HandlerFunc {
	openaiClient := do.MustInvoke[*openai.Client](i)

	return func(c *gin.Context) {
		var req TranslationHandlerRequest

		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		prompt := strings.Replace(promptTemplate, "${characters}", req.Characters, 1)

		resp, err := openaiClient.CreateChatCompletion(
			c.Request.Context(),
			openai.ChatCompletionRequest{
				Model: openai.GPT4,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: prompt,
					},
				},
			},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var translationResponse TranslationHandlerResponse

		err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &translationResponse)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse GPT response"})
			return
		}

		translationResponse.Characters = req.Characters

		c.JSON(http.StatusOK, translationResponse)
	}
}
