package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
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

func GetHandler(sqsClient *sqs.Client) infra.LambdaHandler {
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

		lowerTxt := strings.ToLower(update.Message.Text)
		if strings.HasPrefix(lowerTxt, "buat cerita ") {
			story := update.Message.Text
			job := map[string]any{
				"task":  "make-story",
				"topic": story,
			}

			jobBytes, err := json.Marshal(job)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal job: %w", err)
			}

			jobStr := string(jobBytes)
			_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
				MessageBody: &jobStr,
				QueueUrl:    aws.String(os.Getenv("SQS_URL")),
			})

			if err != nil {
				return nil, fmt.Errorf("failed to send message to sqs: %w", err)
			}
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
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(os.Getenv("AWS_REGION")))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	sqsClient := sqs.NewFromConfig(cfg)

	handler := GetHandler(sqsClient)
	handler = WrapHandler(handler)
	lambda.Start(handler)
}
