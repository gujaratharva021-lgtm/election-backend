package controllers

import (
	"election_backend/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetInsights(c *gin.Context) {
	yearStr := c.Query("year")
	year := 2024
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}

	insights, err := services.GetInsightsFromDB(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"year":     year,
		"insights": insights,
		"total":    len(insights),
	})
}

func GetConstituencyInsights(c *gin.Context) {
	// BUG FIX: pehle year pass ho raha tha, ab constituency_id pass hoga
	idStr := c.Param("constituency_id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid constituency ID"})
		return
	}

	insights, err := services.GetInsightsByConstituency(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":         true,
		"constituency_id": id,
		"insights":        insights,
		"total":           len(insights),
	})
}

func GetVoteShareTrend(c *gin.Context) {
	partyIDStr := c.Param("party_id")
	partyID, err := strconv.ParseUint(partyIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid party ID"})
		return
	}

	trends, err := services.GetVoteShareTrend(uint(partyID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"party_id": partyID,
		"trends":   trends,
		"total":    len(trends),
	})
}
