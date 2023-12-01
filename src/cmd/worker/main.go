package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/niyoko/family-assistant/src/infra"
	openai "github.com/sashabaranov/go-openai"

	_ "embed"
)

//go:embed system.txt
var system string

func respondWithError(status int, message string) ([]byte, error) {
	body, err := json.Marshal(&struct {
		Message string `json:"message"`
	}{Message: message})

	if err != nil {
		return nil, fmt.Errorf("failed to marshall body: %w", err)
	}

	resp := &infra.Response{
		StatusCode:      status,
		Body:            string(body),
		IsBase64Encoded: false,
	}

	respJson, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshall response: %w", err)
	}

	return respJson, nil
}

func GetHandler(bot *tgbotapi.BotAPI, openaiClient *openai.Client) infra.LambdaHandler {
	return func(ctx context.Context, payload []byte) ([]byte, error) {
		return json.Marshal(map[string]bool{"ok": true})
	}
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	cfg := openai.DefaultConfig(os.Getenv("OPENAI_KEY"))
	cfg.OrgID = os.Getenv("OPENAI_ORG")
	client := openai.NewClientWithConfig(cfg)

	handler := GetHandler(bot, client)
	handler = infra.WrapHandler(handler)
	lambda.Start(handler)
}
