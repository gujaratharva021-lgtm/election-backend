// controllers/result_controller.go
package controllers

import (
	"election_backend/config"
	"election_backend/models"
	"election_backend/repository"
	"election_backend/scrapers"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetResults(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", "2024"))
	state := c.DefaultQuery("state", "Maharashtra")
	repo := repository.NewResultRepository(config.DB)
	results, err := repo.GetAll(year, state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"year": year, "state": state, "count": len(results), "results": results})
}

func GetVoteTrends(c *gin.Context) {
	state := c.DefaultQuery("state", "Maharashtra")
	repo := repository.NewResultRepository(config.DB)
	trends, err := repo.GetVoteTrends(state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"state": state, "trends": trends})
}

func GetPartyWiseSummary(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", "2024"))
	state := c.DefaultQuery("state", "Maharashtra")
	electionType := c.DefaultQuery("type", "")
	repo := repository.NewResultRepository(config.DB)
	summary, err := repo.GetPartyWiseSummary(year, state, electionType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"year": year, "state": state, "type": electionType, "summary": summary})
}

func GetWinners(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", "2024"))
	state := c.DefaultQuery("state", "Maharashtra")
	electionType := c.DefaultQuery("type", "")
	repo := repository.NewResultRepository(config.DB)
	winners, err := repo.GetWinners(year, state, electionType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"year": year, "state": state, "type": electionType, "count": len(winners), "winners": winners})
}

func FixWinnersFromCSV(c *gin.Context) {
	err := config.DB.Exec(`
UPDATE results SET is_winner = true
WHERE id IN (
SELECT DISTINCT ON (constituency_id) id
FROM results
WHERE election_year = 2024
AND candidate_id NOT IN (
SELECT id FROM candidates WHERE name = 'NOTA'
)
ORDER BY constituency_id, votes DESC
)
`).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Winners updated!"})
}

func GetKhasdars(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", "2024"))
	type KhasdarRow struct {
		PCNo       string  `json:"pc_no"`
		PCName     string  `json:"pc_name"`
		Candidate  string  `json:"candidate"`
		Party      string  `json:"party"`
		TotalVotes int     `json:"total_votes"`
		VoteShare  float64 `json:"vote_share"`
	}
	var khasdars []KhasdarRow
	err := config.DB.Raw(`
SELECT 
c.id::text as pc_no,
c.name as pc_name,
ca.name as candidate,
p.name as party,
r.votes as total_votes,
r.vote_percent as vote_share
FROM results r
JOIN constituencies c ON c.id = r.constituency_id
JOIN candidates ca ON ca.id = r.candidate_id
JOIN parties p ON p.id = r.party_id
WHERE c.type = 'PC'
AND r.is_winner = true
AND r.election_year = ?
`, year).Scan(&khasdars).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"year": 2024, "state": "Maharashtra", "count": len(khasdars), "khasdars": khasdars})
}

func ScrapeECIResults(c *gin.Context) {
	yearStr := c.Query("year")
	if yearStr == "" {
		yearStr = c.PostForm("year")
	}
	if yearStr == "" {
		yearStr = "2024"
	}
	year, _ := strconv.Atoi(yearStr)
	state := c.Query("state")
	if state == "" {
		state = c.PostForm("state")
	}
	if state == "" {
		state = "Maharashtra"
	}
	scraper := scrapers.NewECIScraper()
	results, err := scraper.ScrapeResults(year, state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	repo := repository.NewResultRepository(config.DB)
	if err := repo.SaveResults(results); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB save failed: " + err.Error()})
		return
	}
	config.DB.Create(&models.ScrapeLog{Source: "ECI Results - Datameet", Status: "success", Records: len(results)})
	c.JSON(http.StatusOK, gin.H{"message": "ECI Results scraped!", "year": year, "state": state, "records": len(results)})
}

func ScrapeECICandidates(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", "2024"))
	state := c.DefaultQuery("state", "Maharashtra")
	scraper := scrapers.NewECIScraper()
	candidates, err := scraper.ScrapeCandidates(year, state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	config.DB.Create(&models.ScrapeLog{Source: "ECI Candidates", Status: "success", Records: len(candidates)})
	c.JSON(http.StatusOK, gin.H{"message": "Candidates scraped!", "candidates": len(candidates)})
}
