// controllers/result_controller.go
package controllers

import (
	"election_backend/config"
	"election_backend/models"
	"election_backend/repository"
	"election_backend/scrapers"
	"encoding/csv"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ── GET ALL RESULTS ───────────────────────────────────
func GetResults(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", "2024"))
	state := c.DefaultQuery("state", "Maharashtra")

	repo := repository.NewResultRepository(config.DB)
	results, err := repo.GetAll(year, state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"year":    year,
		"state":   state,
		"count":   len(results),
		"results": results,
	})
}

// ── GET VOTE TRENDS ───────────────────────────────────
func GetVoteTrends(c *gin.Context) {
	state := c.DefaultQuery("state", "Maharashtra")

	repo := repository.NewResultRepository(config.DB)
	trends, err := repo.GetVoteTrends(state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"state":  state,
		"trends": trends,
	})
}

// ── GET PARTY WISE SUMMARY ────────────────────────────
func GetPartyWiseSummary(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", "2024"))
	state := c.DefaultQuery("state", "Maharashtra")

	repo := repository.NewResultRepository(config.DB)
	summary, err := repo.GetPartyWiseSummary(year, state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"year":    year,
		"state":   state,
		"summary": summary,
	})
}

// ── GET WINNERS (Elected Amdars) ─────────────────────────────
func GetWinners(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", "2024"))
	state := c.DefaultQuery("state", "Maharashtra")

	repo := repository.NewResultRepository(config.DB)
	winners, err := repo.GetWinners(year, state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"year":    year,
		"state":   state,
		"count":   len(winners),
		"winners": winners,
	})
}

// ── FIX WINNERS VIA SQL ───────────────────────────────────────
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

// ── GET KHASDARS (Elected MPs) ────────────────────────────────
func GetKhasdars(c *gin.Context) {
	type KhasdarRow struct {
		PCNo       string  `json:"pc_no"`
		PCName     string  `json:"pc_name"`
		Candidate  string  `json:"candidate"`
		Party      string  `json:"party"`
		TotalVotes int     `json:"total_votes"`
		VoteShare  float64 `json:"vote_share"`
	}

	file, err := os.Open("lok_sabha_2024.csv")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "CSV not found"})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Read() // skip header

	var khasdars []KhasdarRow
	records, _ := reader.ReadAll()
	for _, rec := range records {
		if len(rec) < 7 || rec[6] != "1" {
			continue
		}
		votes, _ := strconv.Atoi(strings.TrimSpace(rec[4]))
		share, _ := strconv.ParseFloat(strings.TrimSpace(rec[5]), 64)
		khasdars = append(khasdars, KhasdarRow{
			PCNo:       rec[0],
			PCName:     strings.TrimSpace(rec[1]),
			Candidate:  strings.TrimSpace(rec[2]),
			Party:      strings.TrimSpace(rec[3]),
			TotalVotes: votes,
			VoteShare:  share,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"year":     2024,
		"state":    "Maharashtra",
		"count":    len(khasdars),
		"khasdars": khasdars,
	})
}

// ── SCRAPE ECI RESULTS ────────────────────────────────
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "DB save failed: " + err.Error(),
		})
		return
	}

	config.DB.Create(&models.ScrapeLog{
		Source:  "ECI Results - Datameet",
		Status:  "success",
		Records: len(results),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "✅ ECI Results scraped successfully!",
		"year":    year,
		"state":   state,
		"records": len(results),
	})
}

// ── SCRAPE ECI CANDIDATES ─────────────────────────────
func ScrapeECICandidates(c *gin.Context) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", "2024"))
	state := c.DefaultQuery("state", "Maharashtra")

	scraper := scrapers.NewECIScraper()
	candidates, err := scraper.ScrapeCandidates(year, state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	config.DB.Create(&models.ScrapeLog{
		Source:  "ECI Candidates",
		Status:  "success",
		Records: len(candidates),
	})

	c.JSON(http.StatusOK, gin.H{
		"message":    "Candidates scraped successfully!",
		"candidates": len(candidates),
	})
}
