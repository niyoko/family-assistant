package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/niyoko/family-assistant/src/utils"
	"github.com/samber/lo"
	openai "github.com/sashabaranov/go-openai"

	_ "embed"
)

//go:embed system.txt
var system string

const MAX_LEN = 4075

type processor struct {
	bot          *tgbotapi.BotAPI
	openaiClient *openai.Client
}

func (p *processor) ProcessTask(ctx context.Context, bag map[string]any) {
	task, taskOk := bag["task"].(string)
	if !taskOk {
		fmt.Printf("failed to get task")
		return
	}

	if task == "make-story" {
		topic, topicOk := bag["topic"].(string)
		if !topicOk {
			fmt.Printf("failed to get topic")
			return
		}

		p.MakeStory(ctx, topic)
	}
}

func (p *processor) sendTypingActionEveryThreeSeconds(ctx context.Context, chatID int64) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.bot.Send(tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping))
		}
	}
}

func (p *processor) MakeStory(ctx context.Context, topic string) {
	chatID, err := strconv.ParseInt(os.Getenv("CHAT_ID"), 10, 64)
	if err != nil {
		fmt.Printf("failed to parse chat id: %v", err)
		return
	}

	req := openai.ChatCompletionRequest{
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: system,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: topic,
			},
		},
		Temperature: 1,
		Model:       openai.GPT4TurboPreview,
	}

	typingCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go p.sendTypingActionEveryThreeSeconds(typingCtx, chatID)
	resp, err := p.openaiClient.CreateChatCompletion(ctx, req)
	if err != nil {
		fmt.Printf("failed to create completion: %v", err)
		return
	}

	content := resp.Choices[0].Message.Content
	lines := strings.Split(content, "\n")
	lines = lo.Filter(lines, func(line string, _ int) bool { return line != "" })

	lineChunks := utils.ChunkLinesToMaxLen(lines, MAX_LEN)
	for _, chunk := range lineChunks {
		_, err := p.bot.Send(tgbotapi.NewMessage(chatID, strings.Join(chunk, "\n")))
		if err != nil {
			fmt.Printf("failed to send message: %v", err)
			return
		}
	}
}
