package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

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
		var decodedPayload *infra.GatewayEvent
		err := json.Unmarshal(payload, &decodedPayload)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		if decodedPayload.Headers["x-telegram-bot-api-secret-token"] != os.Getenv("WEBHOOK_SECRET_TOKEN") {
			return respondWithError(401, "Unauthorized")
		}

		var update tgbotapi.Update
		err = json.Unmarshal([]byte(decodedPayload.Body), &update)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal update: %w", err)
		}

		if update.Message == nil {
			return nil, fmt.Errorf("Not message update")
		}

		if strings.HasPrefix(update.Message.Text, "Buat cerita tentang ") {
			story := update.Message.Text
			story = strings.TrimSuffix(story, ".")
			story = strings.TrimSpace(story)

			req := openai.ChatCompletionRequest{
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleSystem,
						Content: system,
					},
					{
						Role:    openai.ChatMessageRoleUser,
						Content: story,
					},
				},
				Temperature: 1,
				Model:       openai.GPT4TurboPreview,
			}
			resp, err := openaiClient.CreateChatCompletion(ctx, req)
			if err != nil {
				return nil, fmt.Errorf("failed to create completion: %w", err)
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, resp.Choices[0].Message.Content)
			_, err = bot.Send(msg)
			if err != nil {
				return nil, fmt.Errorf("failed to send message: %w", err)
			}

			return json.Marshal(map[string]bool{"ok": true})
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		_, err = bot.Send(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to send message: %w", err)
		}

		return json.Marshal(map[string]bool{"ok": true})
	}
}

func WrapHandler(inner infra.LambdaHandler) infra.LambdaHandler {
	return func(ctx context.Context, payload []byte) ([]byte, error) {
		logItem := map[string]interface{}{
			"payload": string(payload),
		}
		out, err := inner.Invoke(ctx, payload)
		if err != nil {
			logItem["error"] = err.Error()
		} else {
			logItem["response"] = string(out)
		}

		j, _ := json.Marshal(logItem)
		log.Println(string(j))
		return out, err
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
	handler = WrapHandler(handler)
	lambda.Start(handler)
}
