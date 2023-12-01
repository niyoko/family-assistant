package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/aws/aws-lambda-go/lambda"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/niyoko/family-assistant/src/infra"
	openai "github.com/sashabaranov/go-openai"
)

func ProcessRecord(ctx context.Context, wg *sync.WaitGroup, record *infra.SQSRecord) {
	defer wg.Done()

	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	cfg := openai.DefaultConfig(os.Getenv("OPENAI_KEY"))
	cfg.OrgID = os.Getenv("OPENAI_ORG")
	client := openai.NewClientWithConfig(cfg)

	p := processor{
		bot:          bot,
		openaiClient: client,
	}

	bag := make(map[string]any)
	err = json.Unmarshal([]byte(record.Body), &bag)
	if err != nil {
		fmt.Printf("failed to unmarshal body: %v", err)
		return
	}

	p.ProcessTask(ctx, bag)
}

func Handler(ctx context.Context, payload []byte) ([]byte, error) {
	var decodedPayload *infra.SQSRecords
	err := json.Unmarshal(payload, &decodedPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	var wg sync.WaitGroup
	for _, record := range decodedPayload.Records {
		wg.Add(1)
		go ProcessRecord(ctx, &wg, record)
	}

	wg.Wait()
	return json.Marshal(map[string]bool{"ok": true})
}

func main() {
	handler := infra.WrapHandler(Handler)
	lambda.Start(handler)
}
