package main

import (
	"election_backend/config"
	"election_backend/controllers"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	config.ConnectDatabase(cfg)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "Election AI Backend Running 🚀"})
	})

	v1 := r.Group("/api/v1")
	{
		v1.GET("/elections", controllers.GetElections)
		v1.GET("/elections/:year", controllers.GetElectionByYear)
		v1.GET("/constituencies", controllers.GetConstituencies)
		v1.GET("/constituencies/:id", controllers.GetConstituencyById)
		v1.GET("/constituencies/:id/results", controllers.GetConstituencyResults)
		v1.GET("/parties", controllers.GetParties)
		v1.GET("/parties/:id", controllers.GetPartyById)
		v1.GET("/parties/compare", controllers.CompareParties)
		v1.GET("/candidates", controllers.GetCandidates)
		v1.GET("/candidates/:id", controllers.GetCandidateById)
		v1.GET("/results", controllers.GetResults)
		v1.GET("/results/trends", controllers.GetVoteTrends)
		v1.GET("/results/winners", controllers.GetWinners)
		v1.GET("/results/khasdars", controllers.GetKhasdars)
		v1.POST("/results/fix-winners", controllers.FixWinnersFromCSV)
		v1.GET("/results/summary", controllers.GetPartyWiseSummary)
		v1.GET("/insights", controllers.GetInsights)
		v1.GET("/insights/:constituency_id", controllers.GetConstituencyInsights)
		v1.POST("/scrape/results", controllers.ScrapeECIResults)
		v1.POST("/scrape/candidates", controllers.ScrapeECICandidates)
		v1.GET("/parties/:id/trend", controllers.GetVoteShareTrend)
	}

	log.Printf("🚀 Server starting on port %s", cfg.Port)
	r.Run(":" + cfg.Port)
}
