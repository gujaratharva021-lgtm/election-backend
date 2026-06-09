// controllers/constituency_controller.go
package controllers

import (
	"election_backend/config"
	"election_backend/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetConstituencies(c *gin.Context) {
	state := c.DefaultQuery("state", "Maharashtra")

	repo := repository.NewConstituencyRepository(config.DB)
	constituencies, err := repo.GetAll(state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"state":          state,
		"count":          len(constituencies),
		"constituencies": constituencies,
	})
}

func GetConstituencyById(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	repo := repository.NewConstituencyRepository(config.DB)
	constituency, err := repo.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Constituency not found"})
		return
	}

	c.JSON(http.StatusOK, constituency)
}

func GetConstituencyResults(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	year, _ := strconv.Atoi(c.DefaultQuery("year", "2024"))

	repo := repository.NewResultRepository(config.DB)
	results, err := repo.GetByConstituency(uint(id), year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"constituency_id": id,
		"year":            year,
		"results":         results,
	})
}
