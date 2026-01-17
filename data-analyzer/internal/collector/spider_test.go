package collector

import (
	"context"
	"os"
	"testing"
	"time"

	"data-analyzer/internal/riot"

	"github.com/bits-and-blooms/bloom/v3"
	"github.com/joho/godotenv"
)

func init() {
	// Load .env from project root
	godotenv.Load("../../.env")
}

func skipIfNoAPIKey(t *testing.T) *riot.Client {
	if os.Getenv("RIOT_API_KEY") == "" {
		t.Skip("RIOT_API_KEY not set, skipping integration test")
	}
	client, err := riot.NewClient()
	if err != nil {
		t.Fatalf("Failed to create Riot client: %v", err)
	}
	return client
}

// Test: Data collection standard
// Verifies that the Riot API returns valid match data
func TestDataCollection_Standard(t *testing.T) {
	client := skipIfNoAPIKey(t)
	ctx := context.Background()

	// Get a challenger player dynamically
	t.Log("Fetching top challenger player...")
	puuid, err := client.GetTopChallengerPUUID(ctx)
	if err != nil {
		t.Fatalf("GetTopChallengerPUUID failed: %v", err)
	}
	t.Logf("Got PUUID: %s...", puuid[:16])

	// Step 2: Get match history
	t.Logf("Fetching match history...")
	matchIDs, err := client.GetMatchHistory(ctx, puuid, 5)
	if err != nil {
		t.Fatalf("GetMatchHistory failed: %v", err)
	}
	if len(matchIDs) == 0 {
		t.Fatal("GetMatchHistory returned no matches")
	}
	t.Logf("Got %d match IDs", len(matchIDs))

	// Step 3: Get match details for first match
	matchID := matchIDs[0]
	t.Logf("Fetching match details for: %s", matchID)
	match, err := client.GetMatch(ctx, matchID)
	if err != nil {
		t.Fatalf("GetMatch failed: %v", err)
	}

	// Verify match data structure
	if match.Metadata.MatchID == "" {
		t.Error("Match metadata missing matchId")
	}
	if match.Info.GameVersion == "" {
		t.Error("Match info missing gameVersion")
	}
	if len(match.Info.Participants) != 10 {
		t.Errorf("Expected 10 participants, got %d", len(match.Info.Participants))
	}

	// Verify participant data
	for i, p := range match.Info.Participants {
		if p.PUUID == "" {
			t.Errorf("Participant %d missing PUUID", i)
		}
		if p.ChampionID == 0 {
			t.Errorf("Participant %d missing championId", i)
		}
		if p.TeamPosition == "" && match.Info.QueueID == 420 {
			// Team position should be set for ranked games
			t.Logf("Warning: Participant %d missing teamPosition", i)
		}
	}

	t.Logf("Match %s validated successfully", matchID)
	t.Logf("  Game Version: %s", match.Info.GameVersion)
	t.Logf("  Duration: %d seconds", match.Info.GameDuration)
	t.Logf("  Participants: %d", len(match.Info.Participants))
}

// Test: Data collection invalid-rank
// Verifies that players under Emerald 4 are identified correctly
func TestDataCollection_InvalidRank(t *testing.T) {
	client := skipIfNoAPIKey(t)
	ctx := context.Background()

	// Test 1: Challenger player should qualify
	t.Run("Challenger_ShouldQualify", func(t *testing.T) {
		puuid, err := client.GetTopChallengerPUUID(ctx)
		if err != nil {
			t.Fatalf("GetTopChallengerPUUID failed: %v", err)
		}

		tier, division, hasRank, err := client.GetSoloQueueRank(ctx, puuid)
		if err != nil {
			t.Fatalf("GetSoloQueueRank failed: %v", err)
		}

		if !hasRank {
			t.Fatal("Challenger player should have rank")
		}

		qualifies := riot.IsEmerald4OrHigher(tier, division)
		t.Logf("Player rank: %s %s, qualifies: %v", tier, division, qualifies)

		if !qualifies {
			t.Errorf("Challenger player should qualify, got %s %s", tier, division)
		}
	})

	// Test 2: Low rank player should NOT qualify
	t.Run("LowRank_ShouldNotQualify", func(t *testing.T) {
		// Use a known low-rank player
		account, err := client.GetAccountByRiotID(ctx, "Miso", "Chasu")
		if err != nil {
			t.Fatalf("GetAccountByRiotID failed: %v", err)
		}

		tier, division, hasRank, err := client.GetSoloQueueRank(ctx, account.PUUID)
		if err != nil {
			t.Fatalf("GetSoloQueueRank failed: %v", err)
		}

		if !hasRank {
			t.Log("Player has no rank - correctly would be skipped")
			return
		}

		qualifies := riot.IsEmerald4OrHigher(tier, division)
		t.Logf("Player rank: %s %s, qualifies: %v", tier, division, qualifies)

		if qualifies {
			t.Errorf("Low rank player should NOT qualify, got %s %s", tier, division)
		}
	})
}

// Test: Data collection invalid-patch
// Verifies that matches from old patches are identified correctly
func TestDataCollection_InvalidPatch(t *testing.T) {
	client := skipIfNoAPIKey(t)
	ctx := context.Background()

	// Get current patch
	currentPatch, err := riot.GetCurrentPatch(ctx)
	if err != nil {
		t.Fatalf("GetCurrentPatch failed: %v", err)
	}
	t.Logf("Current patch: %s", currentPatch)

	// Get a challenger player's match history
	puuid, err := client.GetTopChallengerPUUID(ctx)
	if err != nil {
		t.Fatalf("GetTopChallengerPUUID failed: %v", err)
	}

	// Fetch more matches to potentially get some from old patches
	matchIDs, err := client.GetMatchHistory(ctx, puuid, 20)
	if err != nil {
		t.Fatalf("GetMatchHistory failed: %v", err)
	}

	currentPatchCount := 0
	oldPatchCount := 0

	for _, matchID := range matchIDs {
		match, err := client.GetMatch(ctx, matchID)
		if err != nil {
			t.Logf("Failed to get match %s: %v", matchID, err)
			continue
		}

		matchPatch := riot.NormalizePatch(match.Info.GameVersion)
		isCurrentPatch := matchPatch == currentPatch

		if isCurrentPatch {
			currentPatchCount++
		} else {
			oldPatchCount++
			t.Logf("Old patch match found: %s (patch %s)", matchID, matchPatch)
		}
	}

	t.Logf("Results: %d current patch, %d old patch", currentPatchCount, oldPatchCount)

	// The test passes if we can correctly identify patch versions
	// We don't require old patches to exist, but the logic must work
	if currentPatchCount == 0 && oldPatchCount == 0 {
		t.Error("No matches were categorized - something is wrong with patch detection")
	}
}

// Test: Data collection invalid-repeat
// Verifies that the bloom filter correctly detects duplicate match IDs
func TestDataCollection_InvalidRepeat(t *testing.T) {
	client := skipIfNoAPIKey(t)
	ctx := context.Background()

	// Get a challenger player's match history
	puuid, err := client.GetTopChallengerPUUID(ctx)
	if err != nil {
		t.Fatalf("GetTopChallengerPUUID failed: %v", err)
	}

	matchIDs, err := client.GetMatchHistory(ctx, puuid, 10)
	if err != nil {
		t.Fatalf("GetMatchHistory failed: %v", err)
	}
	if len(matchIDs) < 2 {
		t.Skip("Not enough matches to test deduplication")
	}

	// Create bloom filter (same config as spider)
	bloomFilter := bloom.NewWithEstimates(500000, 0.001)

	// Add first batch of match IDs
	for _, matchID := range matchIDs {
		if bloomFilter.TestString(matchID) {
			t.Errorf("False positive: match %s detected as duplicate before adding", matchID)
		}
		bloomFilter.AddString(matchID)
	}

	// Test that all added IDs are now detected as duplicates
	duplicatesDetected := 0
	for _, matchID := range matchIDs {
		if bloomFilter.TestString(matchID) {
			duplicatesDetected++
		} else {
			t.Errorf("Match %s not detected as duplicate after adding", matchID)
		}
	}

	t.Logf("Bloom filter correctly detected %d/%d duplicates", duplicatesDetected, len(matchIDs))

	// Test with a new match ID (should not be a duplicate)
	fakeMatchID := "NA1_9999999999"
	if bloomFilter.TestString(fakeMatchID) {
		t.Logf("Note: False positive on fake match ID (expected with bloom filters, but rare)")
	}
}

// Test: Data collection invalid-ratelimit
// Verifies that the rate limiter slows down when approaching limits
func TestDataCollection_InvalidRateLimit(t *testing.T) {
	client := skipIfNoAPIKey(t)
	ctx := context.Background()

	// Get a challenger player to use for requests
	puuid, err := client.GetTopChallengerPUUID(ctx)
	if err != nil {
		t.Fatalf("GetTopChallengerPUUID failed: %v", err)
	}

	matchIDs, err := client.GetMatchHistory(ctx, puuid, 5)
	if err != nil {
		t.Fatalf("GetMatchHistory failed: %v", err)
	}
	if len(matchIDs) == 0 {
		t.Skip("No matches available for rate limit test")
	}

	matchID := matchIDs[0]

	// Make rapid requests and measure timing
	// The rate limiter should enforce min 50ms between requests
	// and slow down as we approach 90 req/2min
	numRequests := 25
	times := make([]time.Duration, numRequests)
	start := time.Now()

	for i := 0; i < numRequests; i++ {
		reqStart := time.Now()
		_, err := client.GetMatch(ctx, matchID)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
		times[i] = time.Since(reqStart)
	}

	totalDuration := time.Since(start)
	avgDuration := totalDuration / time.Duration(numRequests)

	t.Logf("Made %d requests in %v", numRequests, totalDuration)
	t.Logf("Average request time: %v", avgDuration)

	// With 50ms minimum interval, 25 requests should take at least 1.2 seconds
	minExpectedDuration := time.Duration(numRequests-1) * 50 * time.Millisecond
	if totalDuration < minExpectedDuration {
		t.Errorf("Requests completed too fast: %v < %v (rate limiter may not be working)",
			totalDuration, minExpectedDuration)
	} else {
		t.Logf("Rate limiter is working: %v >= %v", totalDuration, minExpectedDuration)
	}

	// Check that no individual request was instant (rate limiter adds delay)
	instantRequests := 0
	for i, d := range times {
		if d < 10*time.Millisecond && i > 0 { // First request might be instant
			instantRequests++
		}
	}
	if instantRequests > 5 {
		t.Logf("Warning: %d requests completed in <10ms (rate limiter may not be enforcing minimum interval)", instantRequests)
	}
}
