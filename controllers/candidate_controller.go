package controllers

import (
	"election_backend/config"
	"election_backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetCandidates(c *gin.Context) {
	year := c.Query("year")
	constituencyID := c.Query("constituency_id")
	partyID := c.Query("party_id")

	db := config.DB

	var candidates []models.Candidate
	query := db.Preload("Party").Preload("Constituency")

	if year != "" {
		query = query.Where("election_year = ?", year)
	}
	if constituencyID != "" {
		query = query.Where("constituency_id = ?", constituencyID)
	}
	if partyID != "" {
		query = query.Where("party_id = ?", partyID)
	}

	query.Find(&candidates)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"candidates": candidates,
		"total":      len(candidates),
	})
}

func GetCandidateById(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var candidate models.Candidate
	result := config.DB.Preload("Party").Preload("Constituency").First(&candidate, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Candidate not found"})
		return
	}

	var results []models.Result
	config.DB.Where("candidate_id = ?", id).
		Preload("Constituency").
		Find(&results)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"candidate": candidate,
		"history":   results,
	})
}
