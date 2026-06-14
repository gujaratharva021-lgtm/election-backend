// scrapers/eci_scraper.go
package scrapers

import (
	"election_backend/models"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type ECIScraper struct {
	client *http.Client
}

func NewECIScraper() *ECIScraper {
	return &ECIScraper{
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// Datameet ka ek bada CSV — saare assembly elections
const assemblyCSVURL = "https://raw.githubusercontent.com/datameet/india-election-data/master/assembly-elections/assembly.csv"

var stateYearCSV = map[string]map[int]string{
	"maharashtra": {
		2024: "https://data.opencity.in/dataset/0e9001ca-b19d-43e0-b0f5-73ed286bbacc/resource/1b620173-b21b-4a58-8304-93095429fab5/download/a6c740cd-c7c5-4d04-a88f-3835aabbe0fe.csv",
		2019: "https://raw.githubusercontent.com/datameet/india-election-data/master/assembly-elections/assembly.csv",
		2014: "https://raw.githubusercontent.com/datameet/india-election-data/master/assembly-elections/assembly.csv",
		2009: "https://raw.githubusercontent.com/datameet/india-election-data/master/assembly-elections/assembly.csv",
	},
}

// ── SCRAPE RESULTS ────────────────────────────────────
func (s *ECIScraper) ScrapeResults(year int, state string) ([]models.Result, error) {
	log.Printf("🔍 Fetching data for year=%d state=%s", year, state)

	stateLower := strings.ToLower(state)

	// Check state-year specific CSV
	if yearMap, ok := stateYearCSV[stateLower]; ok {
		if csvURL, ok := yearMap[year]; ok {
			log.Printf("📡 Using OpenCity CSV: %s", csvURL)
			records, err := s.fetchCSV(csvURL)
			if err != nil {
				return nil, fmt.Errorf("CSV fetch failed: %w", err)
			}
			log.Printf("📊 Total records: %d", len(records))
			results := s.parseOpenCityCSV(records, year, state)
			log.Printf("✅ Parsed %d results", len(results))
			return results, nil
		}
	}

	// Local assembly.csv fallback
	records, err := s.readLocalCSV("C:/Users/Atharva/assembly.csv")
	if err != nil {
		log.Printf("⚠️ Local CSV failed, trying remote...")
		records, err = s.fetchCSV(assemblyCSVURL)
		if err != nil {
			return nil, fmt.Errorf("CSV fetch failed: %w", err)
		}
	}
	results := s.filterAndParseResults(records, year, state)
	return results, nil
}

func (s *ECIScraper) readLocalCSV(path string) ([][]string, error) {
	log.Printf("📂 Reading local CSV: %s", path)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	log.Printf("📊 Local CSV records: %d", len(records))
	return records, nil
}

// ── FETCH CSV ─────────────────────────────────────────
func (s *ECIScraper) fetchCSV(url string) ([][]string, error) {
	log.Printf("📡 Fetching: %s", url)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	reader := csv.NewReader(resp.Body)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1 // allow variable columns

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("CSV parse error: %w", err)
	}

	return records, nil
}

// ── FILTER & PARSE ────────────────────────────────────
func (s *ECIScraper) filterAndParseResults(records [][]string, year int, state string) []models.Result {
	var results []models.Result

	if len(records) < 2 {
		return results
	}

	headers := records[0]
	colMap := mapColumns(headers)
	log.Printf("📋 CSV Columns: %v", headers)

	stateLower := strings.ToLower(state)

	for _, row := range records[1:] {
		if len(row) < 12 {
			continue
		}

		rowState := strings.ToLower(strings.TrimSpace(row[0]))
		if !strings.Contains(rowState, stateLower) {
			continue
		}

		rowYear, _ := strconv.Atoi(strings.TrimSpace(row[1]))
		if year != 0 && rowYear != year {
			continue
		}

		acNo := strings.TrimSpace(row[2])
		position := strings.TrimSpace(row[3])
		acName := strings.TrimSpace(row[4])
		candidateName := strings.TrimSpace(getCol(row, colMap, "name"))
		sex := strings.TrimSpace(getCol(row, colMap, "sex"))
		partyName := strings.TrimSpace(getCol(row, colMap, "party"))
		votes := parseInt(getCol(row, colMap, "votes"))
		isWinner := position == "1"

		result := models.Result{
			ElectionYear: rowYear,
			Votes:        votes,
			VotePercent:  0.0,
			IsWinner:     isWinner,
			Candidate: models.Candidate{
				Name:         candidateName,
				ElectionYear: rowYear,
				Gender:       sex,
			},
			Party: models.Party{
				Name:      partyName,
				ShortName: partyName,
			},
			Constituency: models.Constituency{
				Name:  acName,
				State: state,
				Code:  acNo,
			},
		}

		results = append(results, result)
	}

	return results
}

// ── SCRAPE CANDIDATES ─────────────────────────────────
func (s *ECIScraper) ScrapeCandidates(year int, state string) ([]models.Candidate, error) {
	log.Printf("🔍 Fetching candidates from assembly.csv...")

	records, err := s.fetchCSV(assemblyCSVURL)
	if err != nil {
		return nil, fmt.Errorf("CSV fetch failed: %w", err)
	}

	candidates := s.filterAndParseCandidates(records, year, state)
	log.Printf("✅ Filtered %d candidates", len(candidates))
	return candidates, nil
}

// ── PARSE CANDIDATES ──────────────────────────────────
func (s *ECIScraper) filterAndParseCandidates(records [][]string, year int, state string) []models.Candidate {
	var candidates []models.Candidate

	if len(records) < 2 {
		return candidates
	}

	headers := records[0]
	colMap := mapColumns(headers)
	stateLower := strings.ToLower(state)

	for _, row := range records[1:] {
		if len(row) < 5 {
			continue
		}

		rowState := strings.ToLower(getCol(row, colMap, "state_name", "state", "st_name"))
		if !strings.Contains(rowState, stateLower) {
			continue
		}

		rowYear, _ := strconv.Atoi(getCol(row, colMap, "year", "election_year"))
		if year != 0 && rowYear != year {
			continue
		}

		name := getCol(row, colMap, "candidate", "cand_name", "candidate_name")
		if name == "" {
			continue
		}

		candidate := models.Candidate{
			Name:         name,
			ElectionYear: rowYear,
			Gender:       getCol(row, colMap, "gender", "sex"),
		}

		candidates = append(candidates, candidate)
	}

	return candidates
}

// ── PARSE OPENCITY CSV ────────────────────────────────
func (s *ECIScraper) parseOpenCityCSV(records [][]string, year int, state string) []models.Result {
	var results []models.Result

	if len(records) < 2 {
		return results
	}

	headers := records[0]
	colMap := mapColumns(headers)
	log.Printf("📋 OpenCity Columns: %v", headers)

	// Track winners per constituency (highest votes)
	type constEntry struct {
		acNo     string
		maxVotes int
	}
	acWinners := make(map[string]int) // acNo -> maxVotes

	// First pass: find max votes per constituency
	for _, row := range records[1:] {
		if len(row) < 8 {
			continue
		}
		acNo := getCol(row, colMap, "ac no", "ac_no")
		votes := parseInt(getCol(row, colMap, "total votes", "total_votes"))
		if current, ok := acWinners[acNo]; !ok || votes > current {
			acWinners[acNo] = votes
		}
	}

	// Second pass: parse results
	for _, row := range records[1:] {
		if len(row) < 8 {
			continue
		}

		acNo := getCol(row, colMap, "ac no", "ac_no")
		votes := parseInt(getCol(row, colMap, "total votes", "total_votes"))
		voteShare := parseFloat(getCol(row, colMap, "vote share (%)", "vote_share_(%)", "vote share"))
		isWinner := acWinners[acNo] == votes

		acName := getCol(row, colMap, "ac name", "ac_name")
		candidate := getCol(row, colMap, "candidate")
		party := getCol(row, colMap, "party")

		result := models.Result{
			ElectionYear: year,
			Votes:        votes,
			VotePercent:  voteShare,
			IsWinner:     isWinner,
			Candidate: models.Candidate{
				Name:         candidate,
				ElectionYear: year,
			},
			Party: models.Party{
				Name:      party,
				ShortName: party,
			},
			Constituency: models.Constituency{
				Name:  acName,
				State: state,
				Code:  getCol(row, colMap, "ac no", "ac_no"),
			},
		}

		results = append(results, result)
	}

	return results
}

// ── HELPERS ───────────────────────────────────────────
func mapColumns(headers []string) map[string]int {
	m := make(map[string]int)
	for i, h := range headers {
		m[strings.ToLower(strings.TrimSpace(h))] = i
	}
	return m
}

func getCol(row []string, colMap map[string]int, keys ...string) string {
	for _, key := range keys {
		if idx, ok := colMap[key]; ok && idx < len(row) {
			return strings.TrimSpace(row[idx])
		}
	}
	return ""
}

func parseInt(s string) int {
	s = strings.ReplaceAll(s, ",", "")
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}

func (s *ECIScraper) DebugCSV() ([]string, [][]string, error) {
	records, err := s.fetchCSV(assemblyCSVURL)
	if err != nil {
		return nil, nil, err
	}
	if len(records) == 0 {
		return nil, nil, fmt.Errorf("empty CSV")
	}
	headers := records[0]
	sample := [][]string{}
	for i := 1; i <= 3 && i < len(records); i++ {
		sample = append(sample, records[i])
	}
	return headers, sample, nil
}
