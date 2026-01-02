package main

import (
	"fmt"
	"strings"

	"ghostdraft/internal/lcu"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// TeamCompData holds analyzed team composition data
type TeamCompData struct {
	Tags      map[string]int
	AP        int
	AD        int
	Archetype string
	HasTank   bool
	HasPick   bool // Single-target CC (hooks, roots)
}

// analyzeTeamComp checks team damage balance and emits recommendation
func (a *App) analyzeTeamComp(session *lcu.ChampSelectSession, localChampID int) {
	if a.championDB == nil {
		fmt.Println("Team comp: championDB is nil!")
		return
	}
	fmt.Println("Analyzing team comp...")

	var apCount, adCount, mixedCount int
	localHasLocked := false

	for _, player := range session.MyTeam {
		// Skip players without a champion
		if player.ChampionID == 0 {
			continue
		}

		// Skip local player - we want to advise them, not count their hover
		if player.CellID == session.LocalPlayerCellID {
			// Check if local player has locked
			for _, actionGroup := range session.Actions {
				for _, action := range actionGroup {
					if action.ActorCellID == player.CellID && action.Type == "pick" && action.Completed {
						localHasLocked = true
						break
					}
				}
				if localHasLocked {
					break
				}
			}
			continue
		}

		// Count teammate's champion damage type
		champName := a.champions.GetName(player.ChampionID)
		dmgType := a.championDB.GetDamageType(champName)

		fmt.Printf("  Teammate %s: %s\n", champName, dmgType)

		switch dmgType {
		case "AP":
			apCount++
		case "AD":
			adCount++
		default:
			// Mixed types like AD/AP, AP/Tank count as 0.5 each
			if strings.Contains(dmgType, "AP") {
				apCount++
			}
			if strings.Contains(dmgType, "AD") {
				adCount++
			}
			if dmgType == "Tank" {
				mixedCount++
			}
		}
	}

	totalDmgChamps := apCount + adCount
	fmt.Printf("Team comp analysis: AP=%d, AD=%d, Mixed=%d, LocalLocked=%v\n", apCount, adCount, mixedCount, localHasLocked)

	// Don't show recommendation if local player already locked
	if localHasLocked {
		runtime.EventsEmit(a.ctx, "teamcomp:update", map[string]interface{}{
			"show": false,
		})
		return
	}

	// Need at least 1 teammate to assess balance
	if totalDmgChamps < 1 {
		runtime.EventsEmit(a.ctx, "teamcomp:update", map[string]interface{}{
			"show": false,
		})
		return
	}

	var recommendation string
	var severity string // "warning" or "critical"

	apRatio := float64(apCount) / float64(totalDmgChamps)
	adRatio := float64(adCount) / float64(totalDmgChamps)

	if apRatio >= 0.75 {
		recommendation = "Team is AP heavy - consider picking AD"
		if apRatio >= 0.9 {
			severity = "critical"
		} else {
			severity = "warning"
		}
	} else if adRatio >= 0.75 {
		recommendation = "Team is AD heavy - consider picking AP"
		if adRatio >= 0.9 {
			severity = "critical"
		} else {
			severity = "warning"
		}
	}

	if recommendation != "" {
		fmt.Printf("Team comp: AP=%d, AD=%d, Mixed=%d - %s\n", apCount, adCount, mixedCount, recommendation)
		runtime.EventsEmit(a.ctx, "teamcomp:update", map[string]interface{}{
			"show":           true,
			"recommendation": recommendation,
			"severity":       severity,
			"apCount":        apCount,
			"adCount":        adCount,
		})
	} else {
		runtime.EventsEmit(a.ctx, "teamcomp:update", map[string]interface{}{
			"show": false,
		})
	}
}

// analyzeFullComp analyzes both teams when all players have locked in
func (a *App) analyzeFullComp(session *lcu.ChampSelectSession) {
	if a.championDB == nil {
		return
	}

	// Check if all players have locked in
	allLocked := true
	for _, player := range session.MyTeam {
		if player.ChampionID == 0 {
			allLocked = false
			break
		}
	}
	for _, player := range session.TheirTeam {
		if player.ChampionID == 0 {
			allLocked = false
			break
		}
	}

	if !allLocked {
		runtime.EventsEmit(a.ctx, "fullcomp:update", map[string]interface{}{
			"ready": false,
		})
		return
	}

	// Analyze both teams
	allyComp := a.analyzeTeamTags(session.MyTeam)
	enemyComp := a.analyzeTeamTags(session.TheirTeam)

	// Calculate damage percentages
	allyTotal := allyComp.AP + allyComp.AD
	enemyTotal := enemyComp.AP + enemyComp.AD
	allyAPPct, allyADPct := 50, 50
	enemyAPPct, enemyADPct := 50, 50

	if allyTotal > 0 {
		allyAPPct = allyComp.AP * 100 / allyTotal
		allyADPct = 100 - allyAPPct
	}
	if enemyTotal > 0 {
		enemyAPPct = enemyComp.AP * 100 / enemyTotal
		enemyADPct = 100 - enemyAPPct
	}

	fmt.Printf("Full comp: Ally=%s (AP=%d%% AD=%d%%), Enemy=%s (AP=%d%% AD=%d%%)\n",
		allyComp.Archetype, allyAPPct, allyADPct, enemyComp.Archetype, enemyAPPct, enemyADPct)

	runtime.EventsEmit(a.ctx, "fullcomp:update", map[string]interface{}{
		"ready":          true,
		"allyArchetype":  allyComp.Archetype,
		"allyTags":       formatTagCounts(allyComp.Tags),
		"allyAP":         allyAPPct,
		"allyAD":         allyADPct,
		"enemyArchetype": enemyComp.Archetype,
		"enemyTags":      formatTagCounts(enemyComp.Tags),
		"enemyAP":        enemyAPPct,
		"enemyAD":        enemyADPct,
	})
}

// analyzeTeamTags analyzes a team's composition
func (a *App) analyzeTeamTags(team []lcu.ChampSelectPlayer) TeamCompData {
	comp := TeamCompData{
		Tags: make(map[string]int),
	}

	for _, player := range team {
		if player.ChampionID == 0 {
			continue
		}

		champName := a.champions.GetName(player.ChampionID)
		info, _ := a.championDB.GetChampion(champName)
		if info == nil {
			continue
		}

		// Count damage type
		if strings.Contains(info.DamageType, "AP") {
			comp.AP++
		}
		if strings.Contains(info.DamageType, "AD") {
			comp.AD++
		}
		if info.DamageType == "Tank" {
			comp.HasTank = true
		}

		// Count role tags
		tags := info.RoleTags
		for _, tag := range strings.Split(tags, ", ") {
			tag = strings.TrimSpace(tag)
			// Normalize tags - extract base tag
			baseTag := tag
			if strings.Contains(tag, "(") {
				baseTag = strings.TrimSpace(tag[:strings.Index(tag, "(")])
			}
			if baseTag != "" {
				comp.Tags[baseTag]++
			}

			// Check for pick potential (single-target CC)
			if strings.Contains(tag, "Pick") {
				comp.HasPick = true
			}
			if baseTag == "Tank" {
				comp.HasTank = true
			}
		}
	}

	// Determine archetype
	comp.Archetype = determineArchetype(comp)

	return comp
}

// determineArchetype determines the team's primary archetype
func determineArchetype(comp TeamCompData) string {
	engageCount := comp.Tags["Engage"]
	pokeCount := comp.Tags["Poke"]
	burstCount := comp.Tags["Burst"]
	tankCount := comp.Tags["Tank"]
	bruiserCount := comp.Tags["Bruiser"]
	disengageCount := comp.Tags["Disengage"]

	// Hard Engage: 3+ Engage, usually has Tank/Bruiser
	if engageCount >= 3 && (comp.HasTank || bruiserCount >= 1) {
		return "Hard Engage"
	}

	// Poke/Siege: 3+ Poke, lacks hard engage or has disengage
	if pokeCount >= 3 && (engageCount < 2 || disengageCount >= 1) {
		return "Poke/Siege"
	}

	// Pick Comp: 3+ Burst with single-target CC
	if burstCount >= 3 && comp.HasPick {
		return "Pick Comp"
	}

	// Teamfight: Good balance of engage + burst
	if engageCount >= 2 && burstCount >= 2 {
		return "Teamfight"
	}

	// Skirmish/Split: Bruiser heavy, less teamfight
	if bruiserCount >= 3 {
		return "Skirmish"
	}

	// Tank heavy
	if tankCount >= 2 {
		return "Front-to-Back"
	}

	// Default based on highest count
	if pokeCount >= 2 {
		return "Poke"
	}
	if engageCount >= 2 {
		return "Engage"
	}
	if burstCount >= 2 {
		return "Burst"
	}

	return "Mixed"
}

// formatTagCounts formats tags for display
func formatTagCounts(tags map[string]int) []string {
	var result []string
	// Priority order for display
	priority := []string{"Engage", "Burst", "Poke", "Tank", "Bruiser", "Disengage"}

	for _, tag := range priority {
		if count, ok := tags[tag]; ok && count >= 2 {
			result = append(result, fmt.Sprintf("%s (%d)", tag, count))
		}
	}
	return result
}
