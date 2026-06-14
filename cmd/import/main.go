package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Party struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"uniqueIndex"`
	ShortName string
}
type Constituency struct {
	ID    uint `gorm:"primaryKey"`
	Name  string
	Code  string `gorm:"uniqueIndex"`
	State string
	Type  string
}
type Candidate struct {
	ID             uint
	Name           string
	PartyID        uint
	ConstituencyID uint
	ElectionYear   int
}
type Result struct {
	ID             uint `gorm:"primaryKey"`
	ElectionYear   int
	ConstituencyID uint
	CandidateID    uint
	PartyID        uint
	Votes          int
	IsWinner       bool
}

func main() {
	dsn := "host=awseb-e-2ta8uv25be-stack-awsebrdsdatabase-w9njndnukqjm.c12wueweo2b2.ap-south-1.rds.amazonaws.com user=postgres password=kumbh1234 dbname=ebdb port=5432 sslmode=require"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatal("DB connect failed:", err)
	}

	file, _ := os.Open("loksabha.csv")
	defer file.Close()
	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	records, _ := reader.ReadAll()

	count := 0
	for _, row := range records[1:] {
		if len(row) < 9 {
			continue
		}
		state := strings.TrimSpace(row[1])
		if !strings.Contains(strings.ToLower(state), "maharashtra") {
			continue
		}

		year, _ := strconv.Atoi(strings.TrimSpace(row[0]))
		pcName := strings.TrimSpace(row[2]) + " (LS)"
		name := strings.TrimSpace(row[3])
		partyName := strings.TrimSpace(row[5])
		votes, _ := strconv.Atoi(strings.TrimSpace(row[8]))

		var p Party
		db.Where("name = ?", partyName).First(&p)
		if p.ID == 0 {
			p = Party{Name: partyName, ShortName: partyName}
			db.Create(&p)
		}

		var c Constituency
		db.Where("name = ? AND state = ?", pcName, state).First(&c)
		if c.ID == 0 {
			c = Constituency{Name: pcName, Code: pcName, State: state, Type: "PC"}
			db.Create(&c)
		}

		candidate := Candidate{Name: name, PartyID: p.ID, ConstituencyID: c.ID, ElectionYear: year}
		db.Create(&candidate)
		result := Result{ElectionYear: year, ConstituencyID: c.ID, CandidateID: candidate.ID, PartyID: p.ID, Votes: votes}
		db.Create(&result)
		count++
	}

	db.Exec(`UPDATE results SET is_winner = true WHERE id IN (
		SELECT DISTINCT ON (r.constituency_id) r.id FROM results r
		JOIN constituencies c ON c.id = r.constituency_id
		WHERE c.type = 'PC' ORDER BY r.constituency_id, r.votes DESC)`)

	fmt.Printf("✅ Lok Sabha AWS imported: %d records\n", count)
}
