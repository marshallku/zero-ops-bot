package services

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type N8nClient struct {
    webhookURL    string
    webhookSecret string
    httpClient    *http.Client
}

type WebhookPayload struct {
    Type      string `json:"type"`
    Command   string `json:"command,omitempty"`
    Content   string `json:"content,omitempty"`
    UserID    string `json:"user_id"`
    UserName  string `json:"user_name,omitempty"`
    ChannelID string `json:"channel_id"`
    MessageID string `json:"message_id,omitempty"`
    ThreadID  string `json:"thread_id,omitempty"`
    Timestamp string `json:"timestamp"`
    Source    string `json:"source"`
}

type WebhookResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message,omitempty"`
    Data    any    `json:"data,omitempty"`
}

func NewN8nClient(webhookURL, webhookSecret string) *N8nClient {
    return &N8nClient{
        webhookURL:    webhookURL,
        webhookSecret: webhookSecret,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (c *N8nClient) TriggerWebhook(ctx context.Context, payload WebhookPayload) (*WebhookResponse, error) {
    payload.Timestamp = time.Now().UTC().Format(time.RFC3339)
    payload.Source = "zero-ops-bot"

    body, err := json.Marshal(payload)
    if err != nil {
        return nil, fmt.Errorf("marshal payload: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.webhookURL, bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("create request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")
    if c.webhookSecret != "" {
        req.Header.Set("x-discord-api-key", c.webhookSecret)
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("send request: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
    }

    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("read response: %w", err)
    }

    return &WebhookResponse{
        Success: true,
        Message: string(respBody),
    }, nil
}

func (c *N8nClient) TriggerWebhookAsync(payload WebhookPayload) {
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        c.TriggerWebhook(ctx, payload)
    }()
}
