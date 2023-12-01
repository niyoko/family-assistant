package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
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

func HandleRaw(ctx context.Context, payload []byte) ([]byte, error) {
	var decodedPayload *infra.GatewayEvent
	err := json.Unmarshal(payload, &decodedPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	fmt.Println(string(payload))
	return nil, nil
}

func main() {
	lambda.Start(infra.LambdaHandler(HandleRaw))
}
