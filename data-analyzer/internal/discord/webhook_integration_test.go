package discord

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

// TestWebhookClient_SendKeyExpired_Integration sends a real notification to Discord
func TestWebhookClient_SendKeyExpired_Integration(t *testing.T) {
	godotenv.Load("../../.env")

	webhookURL := os.Getenv("DISCORD_WEBHOOK")
	if webhookURL == "" {
		t.Skip("DISCORD_WEBHOOK not set, skipping integration test")
	}

	client := NewWebhookClient(webhookURL)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := client.SendKeyExpiredNotification(ctx, 47832, 18*time.Hour+32*time.Minute, 2*time.Minute)

	if err != nil {
		t.Fatalf("Failed to send key expired notification: %v", err)
	}

	t.Log("Successfully sent key expired notification to Discord")
}

// TestWebhookClient_SendNewSession_Integration sends a real success notification to Discord
func TestWebhookClient_SendNewSession_Integration(t *testing.T) {
	godotenv.Load("../../.env")

	webhookURL := os.Getenv("DISCORD_WEBHOOK")
	if webhookURL == "" {
		t.Skip("DISCORD_WEBHOOK not set, skipping integration test")
	}

	client := NewWebhookClient(webhookURL)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := client.SendNewSessionNotification(ctx, "RGAPI-test-1234-5678-abcd-efghijklmnop", "Challenger #1")

	if err != nil {
		t.Fatalf("Failed to send new session notification: %v", err)
	}

	t.Log("Successfully sent new session notification to Discord")
}
