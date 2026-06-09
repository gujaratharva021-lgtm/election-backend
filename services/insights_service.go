package services

import (
	"election_backend/config"
	"election_backend/models"
	"fmt"
)

func GenerateInsights(year int) ([]map[string]interface{}, error) {
	db := config.DB
	var insights []map[string]interface{}

	// ── 1. Party wise seats won ──────────────────────────
	type PartyResult struct {
		PartyID    uint
		ShortName  string
		Name       string
		SeatsWon   int
		TotalVotes int
	}

	var partyResults []PartyResult
	db.Raw(`
		SELECT r.party_id, p.short_name, p.name,
		COUNT(CASE WHEN r.is_winner = true THEN 1 END) as seats_won,
		SUM(r.votes) as total_votes
		FROM results r
		JOIN parties p ON p.id = r.party_id
		WHERE r.election_year = ?
		GROUP BY r.party_id, p.short_name, p.name
		ORDER BY seats_won DESC
	`, year).Scan(&partyResults)

	for i, pr := range partyResults {
		insightType := "insight"
		tag := "Party"

		if i == 0 {
			insightType = "strength"
			tag = "Winner"
		} else if i == 1 {
			insightType = "opportunity"
			tag = "Runner Up"
		} else if pr.SeatsWon < 10 {
			insightType = "risk"
			tag = "Low Seats"
		}

		insights = append(insights, map[string]interface{}{
			"type":        insightType,
			"title":       fmt.Sprintf("%s won %d seats", pr.ShortName, pr.SeatsWon),
			"description": fmt.Sprintf("%s secured %d seats with %d total votes in %d election.", pr.Name, pr.SeatsWon, pr.TotalVotes, year),
			"tag":         tag,
			"value":       float64(pr.SeatsWon),
		})
	}

	// ── 2. Highest turnout constituency ─────────────────
	var topConstituency models.Result
	db.Where("election_year = ?", year).
		Order("turnout DESC").
		Preload("Constituency").
		First(&topConstituency)

	if topConstituency.ID != 0 {
		insights = append(insights, map[string]interface{}{
			"type":        "strength",
			"title":       fmt.Sprintf("Highest Turnout: %s", topConstituency.Constituency.Name),
			"description": fmt.Sprintf("%s had highest voter turnout of %.1f%% in %d.", topConstituency.Constituency.Name, topConstituency.Turnout, year),
			"tag":         "Turnout",
			"value":       topConstituency.Turnout,
		})
	}

	// ── 3. Closest margin ───────────────────────────────
	var closestResult models.Result
	db.Where("election_year = ? AND is_winner = true", year).
		Order("margin ASC").
		Preload("Constituency").
		Preload("Party").
		First(&closestResult)

	if closestResult.ID != 0 {
		insights = append(insights, map[string]interface{}{
			"type":        "risk",
			"title":       fmt.Sprintf("Closest Fight: %s", closestResult.Constituency.Name),
			"description": fmt.Sprintf("%s won by only %d votes in %s constituency.", closestResult.Party.ShortName, closestResult.Margin, closestResult.Constituency.Name),
			"tag":         "Close Contest",
			"value":       float64(closestResult.Margin),
		})
	}

	// ── 4. Save to DB — duplicate check ─────────────────
	for _, insight := range insights {
		title := insight["title"].(string)
		var existing models.AIInsight
		result := db.Where("title = ?", title).First(&existing)
		if result.Error != nil {
			// nahi mila, toh create karo
			db.Create(&models.AIInsight{
				Type:         insight["type"].(string),
				Title:        title,
				Description:  insight["description"].(string),
				ElectionYear: year,
			})
		}
	}

	return insights, nil
}

func GetInsightsFromDB(year int) ([]map[string]interface{}, error) {
	db := config.DB
	var dbInsights []models.AIInsight

	// year filter lagao
	db.Where("election_year = ? OR election_year IS NULL", year).Find(&dbInsights)

	var insights []map[string]interface{}
	for _, i := range dbInsights {
		insights = append(insights, map[string]interface{}{
			"type":        i.Type,
			"title":       i.Title,
			"description": i.Description,
			"tag":         i.Type,
			"value":       i.Confidence,
		})
	}

	if len(insights) == 0 {
		return GenerateInsights(year)
	}

	return insights, nil
}

func GetInsightsByConstituency(constituencyID int) ([]map[string]interface{}, error) {
	db := config.DB
	var dbInsights []models.AIInsight

	db.Where("constituency_id = ?", constituencyID).Find(&dbInsights)

	var insights []map[string]interface{}
	for _, i := range dbInsights {
		insights = append(insights, map[string]interface{}{
			"type":        i.Type,
			"title":       i.Title,
			"description": i.Description,
			"tag":         i.Type,
			"value":       i.Confidence,
		})
	}

	// agar kuch nahi mila toh live generate karo
	if len(insights) == 0 {
		db.Raw(`
			SELECT r.party_id, p.short_name, p.name,
			r.votes, r.vote_percent, r.is_winner, r.margin
			FROM results r
			JOIN parties p ON p.id = r.party_id
			WHERE r.constituency_id = ?
			ORDER BY r.votes DESC
		`, constituencyID)

		insights = append(insights, map[string]interface{}{
			"type":        "insight",
			"title":       "No insights yet",
			"description": "Data is being processed for this constituency.",
			"tag":         "Pending",
			"value":       0.0,
		})
	}

	return insights, nil
}

func GetVoteShareTrend(partyID uint) ([]map[string]interface{}, error) {
	db := config.DB

	type TrendRow struct {
		ElectionYear   int
		VoteShare      float64
		SeatsWon       int
		SeatsContested int
		State          string
	}

	var trends []TrendRow
	db.Raw(`
		SELECT election_year, vote_share, seats_won, seats_contested, state
		FROM vote_trends
		WHERE party_id = ?
		ORDER BY election_year ASC
	`, partyID).Scan(&trends)

	var result []map[string]interface{}
	for _, t := range trends {
		result = append(result, map[string]interface{}{
			"year":            t.ElectionYear,
			"vote_share":      t.VoteShare,
			"seats_won":       t.SeatsWon,
			"seats_contested": t.SeatsContested,
			"state":           t.State,
		})
	}

	return result, nil
}
