package infra

import (
	"context"
	"encoding/json"
	"log"
)

type GatewayEvent struct {
	Version         string            `json:"version"`
	RouteKey        string            `json:"routeKey"`
	RawPath         string            `json:"rawPath"`
	RawQuery        string            `json:"rawQueryString"`
	PathParameters  map[string]string `json:"pathParameters"`
	Headers         map[string]string `json:"headers"`
	QueryParameters map[string]string `json:"queryStringParameters"`
	RequestContext  struct {
		ApiId string `json:"apiId"`
		Http  struct {
			Method    string `json:"method"`
			Path      string `json:"path"`
			Protocol  string `json:"protocol"`
			SourceIp  string `json:"sourceIp"`
			UserAgent string `json:"userAgent"`
		} `json:"http"`
	} `json:"requestContext"`
	Body string `json:"body"`
}

type SQSRecord struct {
	MessageId     string `json:"messageId"`
	ReceiptHandle string `json:"receiptHandle"`
	Body          string `json:"body"`
	Attributes    struct {
		ApproximateReceiveCount          string `json:"ApproximateReceiveCount"`
		AWSTraceHeader                   string `json:"AWSTraceHeader"`
		SentTimestamp                    string `json:"SentTimestamp"`
		SenderId                         string `json:"SenderId"`
		ApproximateFirstReceiveTimestamp string `json:"ApproximateFirstReceiveTimestamp"`
	} `json:"attributes"`
	MessageAttributes map[string]interface{} `json:"messageAttributes"`
	Md5OfBody         string                 `json:"md5OfBody"`
	EventSource       string                 `json:"eventSource"`
	EventSourceARN    string                 `json:"eventSourceARN"`
	AwsRegion         string                 `json:"awsRegion"`
}

type SQSRecords struct {
	Records []*SQSRecord `json:"Records"`
}

type Response struct {
	Cookies         []string          `json:"cookies"`
	IsBase64Encoded bool              `json:"isBase64Encoded"`
	StatusCode      int               `json:"statusCode"`
	Headers         map[string]string `json:"headers"`
	Body            string            `json:"body"`
}

type RouteHandler func(ctx context.Context, payload *GatewayEvent) (interface{}, error)

func (handler RouteHandler) Invoke(ctx context.Context, payload *GatewayEvent) (interface{}, error) {
	return handler(ctx, payload)
}

type LambdaHandler func(ctx context.Context, payload []byte) ([]byte, error)

func (handler LambdaHandler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {
	return handler(ctx, payload)
}

func WrapHandler(inner LambdaHandler) LambdaHandler {
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
