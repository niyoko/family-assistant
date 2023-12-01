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
)

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

func GetHandler(bot *tgbotapi.BotAPI) infra.LambdaHandler {
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
			return json.Marshal("")
		}

		return json.Marshal(update.Message.Text)
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

	handler := GetHandler(bot)
	handler = WrapHandler(handler)
	lambda.Start(handler)
}
