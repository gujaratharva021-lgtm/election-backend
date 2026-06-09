package repository

import (
	"election_backend/models"

	"gorm.io/gorm"
)

type PartyRepository struct {
	db *gorm.DB
}

func NewPartyRepository(db *gorm.DB) *PartyRepository {
	return &PartyRepository{db: db}
}

func (r *PartyRepository) GetAll() ([]models.Party, error) {
	var parties []models.Party
	err := r.db.Find(&parties).Error
	return parties, err
}

func (r *PartyRepository) GetByID(id uint) (*models.Party, error) {
	var party models.Party
	err := r.db.First(&party, id).Error
	return &party, err
}

func (r *PartyRepository) GetPartyStats(partyID uint, year int) (map[string]interface{}, error) {
	var results []models.Result
	err := r.db.Where("party_id = ? AND election_year = ?", partyID, year).
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	seatsWon := 0
	totalVotes := 0
	totalVotePercent := 0.0

	for _, r := range results {
		if r.IsWinner {
			seatsWon++
		}
		totalVotes += r.Votes
		totalVotePercent += r.VotePercent
	}

	avgVoteShare := 0.0
	if len(results) > 0 {
		avgVoteShare = totalVotePercent / float64(len(results))
	}

	return map[string]interface{}{
		"party_id":        partyID,
		"year":            year,
		"seats_won":       seatsWon,
		"seats_contested": len(results),
		"total_votes":     totalVotes,
		"avg_vote_share":  avgVoteShare,
	}, nil
}

func (r *PartyRepository) GetVoteTrends(partyID uint) ([]models.VoteTrend, error) {
	var trends []models.VoteTrend
	err := r.db.Where("party_id = ?", partyID).
		Order("election_year ASC").
		Preload("Party").
		Find(&trends).Error
	return trends, err
}
