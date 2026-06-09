// models/models.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// ── ELECTION ──────────────────────────────────────────
type Election struct {
	gorm.Model
	Year        int       `json:"year" gorm:"uniqueIndex"`
	Type        string    `json:"type"` // General, State
	State       string    `json:"state"`
	TotalSeats  int       `json:"total_seats"`
	TotalVoters int64     `json:"total_voters"`
	Turnout     float64   `json:"turnout"`
	Date        time.Time `json:"date"`
}

// ── PARTY ─────────────────────────────────────────────
type Party struct {
	gorm.Model
	Name      string `json:"name" gorm:"uniqueIndex"`
	ShortName string `json:"short_name"`
	Color     string `json:"color"`
	Alliance  string `json:"alliance"` // NDA, INDIA, Other
	LogoURL   string `json:"logo_url"`
	Founded   int    `json:"founded"`
}

// ── CONSTITUENCY ──────────────────────────────────────
type Constituency struct {
	gorm.Model
	Name        string  `json:"name"`
	Code        string  `json:"code" gorm:"uniqueIndex"`
	State       string  `json:"state"`
	Type        string  `json:"type"` // Urban/Rural/Semi-Urban
	TotalVoters int64   `json:"total_voters"`
	ReservedFor string  `json:"reserved_for"` // General/SC/ST
	District    string  `json:"district"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

// ── CANDIDATE ─────────────────────────────────────────
type Candidate struct {
	gorm.Model
	Name            string       `json:"name"`
	PartyID         uint         `json:"party_id"`
	Party           Party        `json:"party" gorm:"foreignKey:PartyID"`
	ConstituencyID  uint         `json:"constituency_id"`
	Constituency    Constituency `json:"constituency" gorm:"foreignKey:ConstituencyID"`
	ElectionYear    int          `json:"election_year"`
	Age             int          `json:"age"`
	Gender          string       `json:"gender"`
	Education       string       `json:"education"`
	CriminalCases   int          `json:"criminal_cases"`
	TotalAssets     int64        `json:"total_assets"`
	Liabilities     int64        `json:"liabilities"`
	ECIAffidavitURL string       `json:"eci_affidavit_url"`
}

// ── RESULT ────────────────────────────────────────────
type Result struct {
	gorm.Model
	ElectionYear   int          `json:"election_year"`
	ConstituencyID uint         `json:"constituency_id"`
	Constituency   Constituency `json:"constituency" gorm:"foreignKey:ConstituencyID"`
	CandidateID    uint         `json:"candidate_id"`
	Candidate      Candidate    `json:"candidate" gorm:"foreignKey:CandidateID"`
	PartyID        uint         `json:"party_id"`
	Party          Party        `json:"party" gorm:"foreignKey:PartyID"`
	Votes          int          `json:"votes"`
	VotePercent    float64      `json:"vote_percent"`
	IsWinner       bool         `json:"is_winner"`
	Margin         int          `json:"margin"`
	TotalVotes     int          `json:"total_votes"`
	Turnout        float64      `json:"turnout"`
	NOTA           int          `json:"nota"`
}

// ── VOTE TREND ────────────────────────────────────────
type VoteTrend struct {
	gorm.Model
	PartyID        uint    `json:"party_id"`
	Party          Party   `json:"party" gorm:"foreignKey:PartyID"`
	ElectionYear   int     `json:"election_year"`
	State          string  `json:"state"`
	VoteShare      float64 `json:"vote_share"`
	SeatsWon       int     `json:"seats_won"`
	SeatsContested int     `json:"seats_contested"`
}

// ── AI INSIGHT ────────────────────────────────────────
type AIInsight struct {
	gorm.Model
	ConstituencyID *uint     `json:"constituency_id"`
	PartyID        *uint     `json:"party_id"`
	ElectionYear   int       `json:"election_year"`
	Type           string    `json:"type"` // strength/opportunity/risk/insight
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Confidence     float64   `json:"confidence"`
	GeneratedAt    time.Time `json:"generated_at"`
}

// ── SCRAPE LOG ────────────────────────────────────────
type ScrapeLog struct {
	gorm.Model
	Source    string    `json:"source"`
	Status    string    `json:"status"` // success/failed
	Records   int       `json:"records"`
	Error     string    `json:"error"`
	ScrapedAt time.Time `json:"scraped_at"`
}
