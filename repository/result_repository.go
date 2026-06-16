// repository/result_repository.go
package repository

import (
	"election_backend/models"

	"gorm.io/gorm"
)

type ResultRepository struct {
	db *gorm.DB
}

func NewResultRepository(db *gorm.DB) *ResultRepository {
	return &ResultRepository{db: db}
}

// ── GET ALL RESULTS ───────────────────────────────────
func (r *ResultRepository) GetAll(year int, state string) ([]models.Result, error) {
	var results []models.Result
	query := r.db.Preload("Candidate").Preload("Party").Preload("Constituency")

	if year != 0 {
		query = query.Where("election_year = ?", year)
	}
	if state != "" {
		query = query.Joins("JOIN constituencies ON constituencies.id = results.constituency_id").
			Where("constituencies.state = ?", state)
	}

	err := query.Find(&results).Error
	return results, err
}

// ── GET WINNERS ONLY ──────────────────────────────────
func (r *ResultRepository) GetWinners(year int, state string, electionType string) ([]models.Result, error) {
	var results []models.Result
	query := r.db.Preload("Candidate").Preload("Party").Preload("Constituency").
		Joins("JOIN constituencies ON constituencies.id = results.constituency_id").
		Where("results.is_winner = true AND results.election_year = ?", year)

	if electionType == "MLA" {
		query = query.Where("constituencies.type = ?", "AC")
	} else if electionType == "MP" {
		query = query.Where("constituencies.type = ?", "PC")
	} else if state != "" {
		query = query.Where("constituencies.state = ?", state)
	}

	err := query.Find(&results).Error
	return results, err
}

// ── GET BY CONSTITUENCY ───────────────────────────────
func (r *ResultRepository) GetByConstituency(constituencyID uint, year int) ([]models.Result, error) {
	var results []models.Result
	err := r.db.Preload("Candidate").Preload("Party").
		Where("constituency_id = ? AND election_year = ?", constituencyID, year).
		Order("votes DESC").
		Find(&results).Error
	return results, err
}

// ── GET VOTE TRENDS ───────────────────────────────────
func (r *ResultRepository) GetVoteTrends(state string) ([]models.VoteTrend, error) {
	var trends []models.VoteTrend
	err := r.db.Preload("Party").
		Where("state = ?", state).
		Order("election_year ASC").
		Find(&trends).Error
	return trends, err
}

// ── SAVE RESULTS ──────────────────────────────────────
func (r *ResultRepository) SaveResults(results []models.Result) error {
	for i := range results {
		// 1. Party pehle save karo
		var party models.Party
		partyName := results[i].Party.Name
		if partyName == "" {
			partyName = "Independent"
		}
		r.db.Where("name = ?", partyName).FirstOrCreate(&party, models.Party{
			Name:      partyName,
			ShortName: partyName,
		})
		results[i].PartyID = party.ID

		// 2. Constituency save karo
		var constituency models.Constituency
		if results[i].Constituency.Name != "" {
			code := results[i].Constituency.Code
			if code == "" {
				code = results[i].Constituency.Name
			}
			r.db.Where("name = ? AND state = ?", results[i].Constituency.Name, results[i].Constituency.State).
				FirstOrCreate(&constituency, models.Constituency{
					Name:  results[i].Constituency.Name,
					State: results[i].Constituency.State,
					Code:  code,
				})
			results[i].ConstituencyID = constituency.ID
		}

		// Agar constituency nahi mili toh default banao
		if results[i].ConstituencyID == 0 {
			var defaultConst models.Constituency
			r.db.Where("name = ? AND state = ?", "Unknown", results[i].Constituency.State).
				FirstOrCreate(&defaultConst, models.Constituency{
					Name:  "Unknown",
					State: "Maharashtra",
					Code:  "UNK",
				})
			results[i].ConstituencyID = defaultConst.ID
		}

		// 3. Candidate save karo — party_id set karke
		var candidate models.Candidate
		candidateName := results[i].Candidate.Name
		if candidateName != "" {
			r.db.Where("name = ? AND election_year = ?", candidateName, results[i].ElectionYear).
				FirstOrCreate(&candidate, models.Candidate{
					Name:           candidateName,
					ElectionYear:   results[i].ElectionYear,
					PartyID:        party.ID,
					ConstituencyID: constituency.ID,
				})
			results[i].CandidateID = candidate.ID
		}

		// 4. Result clear karo nested structs
		results[i].Party = models.Party{}
		results[i].Candidate = models.Candidate{}
		results[i].Constituency = models.Constituency{}
	}

	return r.db.CreateInBatches(results, 100).Error
}

// ── PARTY WISE SUMMARY ────────────────────────────────
func (r *ResultRepository) GetPartyResultSummary(year int, state string, partyID uint) (map[string]interface{}, error) {
	type Summary struct {
		SeatsWon       int
		SeatsContested int
		TotalVotes     int64
		VoteShare      float64
	}

	var s Summary
	r.db.Raw(`
		SELECT
			COUNT(CASE WHEN r.is_winner = true THEN 1 END) as seats_won,
			COUNT(*) as seats_contested,
			SUM(r.votes) as total_votes,
			ROUND(AVG(r.vote_percent), 2) as vote_share
		FROM results r
		JOIN constituencies c ON c.id = r.constituency_id
		WHERE r.election_year = ?
		AND r.party_id = ?
		AND c.state = ?
	`, year, partyID, state).Scan(&s)

	return map[string]interface{}{
		"seats_won":       s.SeatsWon,
		"seats_contested": s.SeatsContested,
		"total_votes":     s.TotalVotes,
		"vote_share":      s.VoteShare,
	}, nil
}

func (r *ResultRepository) GetPartyWiseSummary(year int, state string, electionType string) ([]map[string]interface{}, error) {
	var summary []map[string]interface{}
	query := `
		SELECT 
			p.name as party_name,
			p.short_name,
			p.color,
			COUNT(CASE WHEN r.is_winner THEN 1 END) as seats_won,
			SUM(r.votes) as total_votes,
			ROUND(AVG(r.vote_percent)::numeric, 2) as avg_vote_share
		FROM results r
		JOIN parties p ON p.id = r.party_id
		JOIN constituencies c ON c.id = r.constituency_id
		WHERE r.election_year = ? AND c.state = ?`
	args := []interface{}{year, state}
	if electionType != "" {
		query += ` AND c.type = ?`
		args = append(args, electionType)
	}
	query += ` GROUP BY p.id, p.name, p.short_name, p.color ORDER BY seats_won DESC`
	err := r.db.Raw(query, args...).Scan(&summary).Error
	return summary, err
}
