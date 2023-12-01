package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	openai "github.com/sashabaranov/go-openai"

	_ "embed"
)

//go:embed system.txt
var system string

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

	p.bot.Send(tgbotapi.NewChatAction(chatID, tgbotapi.ChatTyping))
	resp, err := p.openaiClient.CreateChatCompletion(ctx, req)
	if err != nil {
		fmt.Printf("failed to create completion: %v", err)
		return
	}

	content := resp.Choices[0].Message.Content

	p.bot.Send(tgbotapi.NewChatAction(chatID, tgbotapi.ChatRecordVoice))
	voice, err := p.openaiClient.CreateSpeech(ctx, openai.CreateSpeechRequest{
		Model:          openai.TTSModel1,
		Input:          content,
		Voice:          openai.VoiceNova,
		Speed:          0.85,
		ResponseFormat: openai.SpeechResponseFormatOpus,
	})

	if err != nil {
		fmt.Printf("failed to create speech: %v", err)
		return
	}

	defer voice.Close()

	p.bot.Send(tgbotapi.NewChatAction(chatID, tgbotapi.ChatUploadVoice))
	msg := tgbotapi.NewVoice(chatID, tgbotapi.FileReader{
		Name:   "Cerita.oog",
		Reader: voice,
	})
	_, err = p.bot.Send(msg)
	if err != nil {
		fmt.Printf("failed to send message: %v", err)
		return
	}
}
