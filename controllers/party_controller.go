package controllers

import (
	"election_backend/config"
	"election_backend/models"
	"election_backend/repository"
	"election_backend/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetParties(c *gin.Context) {
	var parties []models.Party
	config.DB.Find(&parties)
	c.JSON(http.StatusOK, gin.H{
		"count":   len(parties),
		"parties": parties,
	})
}

func GetPartyById(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var party models.Party
	if err := config.DB.First(&party, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Party not found"})
		return
	}

	c.JSON(http.StatusOK, party)
}

func CompareParties(c *gin.Context) {
	party1Name := c.Query("party1")
	party2Name := c.Query("party2")
	year, _ := strconv.Atoi(c.DefaultQuery("year", "2024"))
	state := c.DefaultQuery("state", "Maharashtra")

	if party1Name == "" || party2Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "party1 and party2 required"})
		return
	}

	// party1 fetch karo
	var p1, p2 models.Party
	if err := config.DB.Where("short_name = ? OR name = ?", party1Name, party1Name).First(&p1).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "party1 not found: " + party1Name})
		return
	}
	if err := config.DB.Where("short_name = ? OR name = ?", party2Name, party2Name).First(&p2).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "party2 not found: " + party2Name})
		return
	}

	repo := repository.NewResultRepository(config.DB)

	// dono parties ka alag alag data fetch karo
	summary1, _ := repo.GetPartyResultSummary(year, state, p1.ID)
	summary2, _ := repo.GetPartyResultSummary(year, state, p2.ID)

	// vote share trend
	trend1, _ := services.GetVoteShareTrend(p1.ID)
	trend2, _ := services.GetVoteShareTrend(p2.ID)

	c.JSON(http.StatusOK, gin.H{
		"year":  year,
		"state": state,
		"party1": gin.H{
			"id":         p1.ID,
			"name":       p1.Name,
			"short_name": p1.ShortName,
			"alliance":   p1.Alliance,
			"color":      p1.Color,
			"summary":    summary1,
			"trend":      trend1,
		},
		"party2": gin.H{
			"id":         p2.ID,
			"name":       p2.Name,
			"short_name": p2.ShortName,
			"alliance":   p2.Alliance,
			"color":      p2.Color,
			"summary":    summary2,
			"trend":      trend2,
		},
	})
}

func GetElections(c *gin.Context) {
	var elections []models.Election
	config.DB.Find(&elections)
	c.JSON(http.StatusOK, gin.H{"elections": elections})
}

func GetElectionByYear(c *gin.Context) {
	year, _ := strconv.Atoi(c.Param("year"))
	var election models.Election
	if err := config.DB.Where("year = ?", year).First(&election).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Election not found"})
		return
	}
	c.JSON(http.StatusOK, election)
}
