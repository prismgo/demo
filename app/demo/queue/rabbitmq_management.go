package queuedemo

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type rabbitMQManagementClient struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

type rabbitMQManagementConsumer struct {
	Queue struct {
		Name  string `json:"name"`
		VHost string `json:"vhost"`
	} `json:"queue"`
	ChannelDetails struct {
		ConnectionName string `json:"connection_name"`
	} `json:"channel_details"`
}

func newRabbitMQManagementClient() (*rabbitMQManagementClient, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("PRISMGO_RABBITMQ_MANAGEMENT_URL")), "/")
	if baseURL == "" {
		baseURL = strings.TrimRight(strings.TrimSpace(os.Getenv("RABBITMQ_MANAGEMENT_URL")), "/")
	}
	if baseURL == "" {
		return nil, fmt.Errorf("queue demo RabbitMQ fault injection requires PRISMGO_RABBITMQ_MANAGEMENT_URL")
	}
	amqpURL := strings.TrimSpace(os.Getenv("PRISMGO_RABBITMQ_TEST_URL"))
	if amqpURL == "" {
		amqpURL = strings.TrimSpace(os.Getenv("RABBITMQ_URL"))
	}
	parsed, err := url.Parse(amqpURL)
	if err != nil || parsed.User == nil {
		return nil, fmt.Errorf("queue demo RabbitMQ fault injection requires credentials in RabbitMQ URL")
	}
	password, _ := parsed.User.Password()
	return &rabbitMQManagementClient{
		baseURL: baseURL, username: parsed.User.Username(), password: password,
		client: &http.Client{Timeout: 5 * time.Second},
	}, nil
}

func (c *rabbitMQManagementClient) connectionForQueue(ctx context.Context, queueName string) (string, error) {
	deadline := time.Now().Add(10 * time.Second)
	for {
		connection, err := c.connectionForQueueOnce(ctx, queueName)
		if err == nil {
			return connection, nil
		}
		if time.Now().After(deadline) {
			return "", err
		}
		timer := time.NewTimer(50 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return "", ctx.Err()
		case <-timer.C:
		}
	}
}

func (c *rabbitMQManagementClient) connectionForQueueOnce(ctx context.Context, queueName string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/consumers/%2F", nil)
	if err != nil {
		return "", fmt.Errorf("build RabbitMQ consumers request: %w", err)
	}
	request.SetBasicAuth(c.username, c.password)
	response, err := c.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("list RabbitMQ consumers: %w", err)
	}
	defer response.Body.Close() //nolint:errcheck // read-only management response; request/decode errors take precedence
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("list RabbitMQ consumers: HTTP %d", response.StatusCode)
	}
	var consumers []rabbitMQManagementConsumer
	if err := json.NewDecoder(response.Body).Decode(&consumers); err != nil {
		return "", fmt.Errorf("decode RabbitMQ consumers: %w", err)
	}
	for _, consumer := range consumers {
		if consumer.Queue.Name == queueName && consumer.Queue.VHost == "/" && consumer.ChannelDetails.ConnectionName != "" {
			return consumer.ChannelDetails.ConnectionName, nil
		}
	}
	return "", fmt.Errorf("RabbitMQ management API returned no live consumer for queue %q", queueName)
}

func (c *rabbitMQManagementClient) closeConnection(ctx context.Context, name string) error {
	endpoint := c.baseURL + "/api/connections/" + url.PathEscape(name)
	request, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("build RabbitMQ close request: %w", err)
	}
	request.SetBasicAuth(c.username, c.password)
	request.Header.Set("X-Reason", "PrismGo reconnect demo")
	response, err := c.client.Do(request)
	if err != nil {
		return fmt.Errorf("close RabbitMQ connection: %w", err)
	}
	defer response.Body.Close() //nolint:errcheck // no response body is consumed after the accepted close operation
	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("close RabbitMQ connection: HTTP %d", response.StatusCode)
	}
	return nil
}

func (c *rabbitMQManagementClient) publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	payload, err := json.Marshal(map[string]any{
		"properties":       map[string]any{},
		"routing_key":      routingKey,
		"payload":          base64.StdEncoding.EncodeToString(body),
		"payload_encoding": "base64",
	})
	if err != nil {
		return fmt.Errorf("encode RabbitMQ publish request: %w", err)
	}
	endpoint := c.baseURL + "/api/exchanges/%2F/" + url.PathEscape(exchange) + "/publish"
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build RabbitMQ publish request: %w", err)
	}
	request.SetBasicAuth(c.username, c.password)
	request.Header.Set("Content-Type", "application/json")
	response, err := c.client.Do(request)
	if err != nil {
		return fmt.Errorf("publish RabbitMQ message: %w", err)
	}
	defer response.Body.Close() //nolint:errcheck // request/decode errors carry the actionable publish context
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("publish RabbitMQ message: HTTP %d", response.StatusCode)
	}
	var result struct {
		Routed bool `json:"routed"`
	}
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode RabbitMQ publish response: %w", err)
	}
	if !result.Routed {
		return fmt.Errorf("publish RabbitMQ message: broker did not route message")
	}
	return nil
}
