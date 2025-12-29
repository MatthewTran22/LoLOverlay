# Matchup Role Fix Plan

## Problem
In `internal/ugg/matchup.go`, the data structure access order is wrong. Currently it accesses:
- `regionMap[roleID]` (line 82) - treating first level as role
- `tierMap["3"]` (line 92) - treating second level as tier

But U.GG's actual structure is **Region -> Tier -> Role**, so "3" at line 92 is selecting ADC role, not Diamond tier.

## Fix

### File: `internal/ugg/matchup.go` (lines 82-95)

**Before:**
```go
roleData, ok := regionMap[roleID]
if !ok {
    continue
}

var tierMap map[string]json.RawMessage
if err := json.Unmarshal(roleData, &tierMap); err != nil {
    continue
}

tierData, ok := tierMap["3"]
```

**After:**
```go
tierData, ok := regionMap["3"]
if !ok {
    continue
}

var roleMap map[string]json.RawMessage
if err := json.Unmarshal(tierData, &roleMap); err != nil {
    continue
}

roleData, ok := roleMap[roleID]
```

## Summary
- Swap the access order: get tier "3" (Diamond+) first, then get role by `roleID`
- Rename variables to match their actual meaning (`tierMap` -> `roleMap`, etc.)
- This makes matchups flexible based on the user's actual role instead of hardcoded to ADC
