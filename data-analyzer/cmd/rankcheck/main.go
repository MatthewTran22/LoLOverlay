package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"data-analyzer/internal/riot"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	godotenv.Load()

	riotID := flag.String("riot-id", "", "Riot ID in format 'GameName#TagLine'")
	flag.Parse()

	if *riotID == "" {
		fmt.Println("Usage: go run cmd/rankcheck/main.go --riot-id=\"PlayerName#NA1\"")
		os.Exit(1)
	}

	// Parse Riot ID
	parts := strings.SplitN(*riotID, "#", 2)
	if len(parts) != 2 {
		log.Fatalf("Invalid Riot ID format. Expected 'GameName#TagLine', got: %s", *riotID)
	}
	gameName, tagLine := parts[0], parts[1]

	// Create Riot client
	client, err := riot.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Riot client: %v", err)
	}

	ctx := context.Background()

	// Step 1: Get account info (PUUID)
	fmt.Printf("\n1. Looking up account: %s#%s\n", gameName, tagLine)
	account, err := client.GetAccountByRiotID(ctx, gameName, tagLine)
	if err != nil {
		log.Fatalf("Failed to get account: %v", err)
	}
	fmt.Printf("   PUUID: %s...%s\n", account.PUUID[:8], account.PUUID[len(account.PUUID)-4:])

	// Step 2: Get ranked entries directly by PUUID
	fmt.Printf("\n2. Getting ranked entries...\n")
	entries, err := client.GetRankedEntriesByPUUID(ctx, account.PUUID)
	if err != nil {
		log.Fatalf("Failed to get ranked entries: %v", err)
	}

	if len(entries) == 0 {
		fmt.Println("   No ranked entries found (unranked)")
	} else {
		for _, entry := range entries {
			queueName := entry.QueueType
			if entry.QueueType == "RANKED_SOLO_5x5" {
				queueName = "Solo/Duo"
			} else if entry.QueueType == "RANKED_FLEX_SR" {
				queueName = "Flex"
			}
			fmt.Printf("   %s: %s %s (%d LP) - %dW %dL\n",
				queueName, entry.Tier, entry.Rank, entry.LeaguePoints, entry.Wins, entry.Losses)
		}
	}

	// Step 3: Test the convenience method
	fmt.Printf("\n3. Testing GetSoloQueueRank()...\n")
	tier, division, hasRank, err := client.GetSoloQueueRank(ctx, account.PUUID)
	if err != nil {
		log.Fatalf("Failed to get solo queue rank: %v", err)
	}

	if !hasRank {
		fmt.Println("   No solo queue rank found")
	} else {
		fmt.Printf("   Solo Queue: %s %s\n", tier, division)
	}

	// Step 4: Check if qualifies for data collection
	fmt.Printf("\n4. Rank filter check...\n")
	if !hasRank {
		fmt.Println("   Result: SKIP (no solo queue rank)")
	} else if riot.IsEmerald4OrHigher(tier, division) {
		fmt.Printf("   Result: PASS (%s %s >= Emerald IV)\n", tier, division)
	} else {
		fmt.Printf("   Result: SKIP (%s %s < Emerald IV)\n", tier, division)
	}

	fmt.Println("\nDone!")
}
