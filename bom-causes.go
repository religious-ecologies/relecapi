package apiary

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
)

// DeathCauses TODO: Describe
type DeathCauses struct {
	DeathID    int        `json:"death_id"`
	Death      string     `json:"death"`
	Count      NullInt64  `json:"count"`
	WeekID     string     `json:"week_id"`
	WeekNo     int        `json:"week_no"`
	StartDay   NullInt64  `json:"start_day"`
	StartMonth NullString `json:"start_month"`
	EndDay     NullInt64  `json:"end_day"`
	EndMonth   NullString `json:"end_month"`
	Year       int        `json:"year"`
	SplitYear  string     `json:"split_year"`
}

// Causes describes a cause of death.
type Causes struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ID          int    `json:"id"`
}

// DeathCausesHandler TODO: Describe
func (s *Server) DeathCausesHandler() http.HandlerFunc {

	queryCause := `
	SELECT 
		c.death_id,
		c.death,
		c.description,
		c.count, 
		c.week_id, 
		w.week_no,
		w.start_day, 
		w.start_month, 
		w.end_day, 
		w.end_month, 
		y.year,
		w.split_year
	FROM 
		bom.causes_of_death c
	JOIN 
		bom.week w ON w.week_id = c.week_id
	JOIN
		bom.year y ON y.year_id = w.year_id
	WHERE 
		y.year >= $1
		AND y.year <= $2
		AND (
			$3::int[] IS NULL
			OR c.death_id = ANY($3::int[])
		)
	ORDER BY 
		y.year ASC,
		w.week_no ASC,
		c.death ASC
	LIMIT $4
	OFFSET $5;
	`

	queryNoCause := `
	SELECT 
		c.death_id, 
		c.death,
		c.description,
		c.count, 
		c.week_id, 
		w.week_no,
		w.start_day, 
		w.start_month, 
		w.end_day, 
		w.end_month, 
		y.year,
		w.split_year
	FROM 
		bom.causes_of_death c
	JOIN 
		bom.week w ON w.week_id = c.week_id
	JOIN
		bom.year y ON y.year_id = w.year_id
	WHERE 
		y.year >= $1
		AND y.year <= $2
	ORDER BY 
		y.year ASC,
		w.week_no ASC,
		c.death ASC
	LIMIT $3
	OFFSET $4;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		startYear := r.URL.Query().Get("start-year")
		endYear := r.URL.Query().Get("end-year")
		causes := r.URL.Query().Get("id")
		limit := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")

		if startYear == "" || endYear == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		startYearInt, err := strconv.Atoi(startYear)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		endYearInt, err := strconv.Atoi(endYear)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// causes needs to be a postgres array of integers to send to the query. This
		// returns '{1, 2, 3}' to give the query a literal array.
		causes = fmt.Sprintf("{%s}", strings.TrimSpace(causes))

		if limit == "" {
			limit = "25"
		}
		if offset == "" {
			offset = "0"
		}

		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		offsetInt, err := strconv.Atoi(offset)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		results := make([]DeathCauses, 0)
		var row DeathCauses
		var rows pgx.Rows

		switch {
		case causes == "{}":
			rows, err = s.DB.Query(context.TODO(), queryNoCause, startYearInt, endYearInt, limitInt, offsetInt)
		case causes != "{}":
			rows, err = s.DB.Query(context.TODO(), queryCause, startYearInt, endYearInt, causes, limitInt, offsetInt)
		default:
			rows, err = s.DB.Query(context.TODO(), queryNoCause, startYearInt, endYearInt, limitInt, offsetInt)
		}

		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(
				&row.DeathID,
				&row.Death,
				&row.Count,
				&row.WeekID,
				&row.WeekNo,
				&row.StartDay,
				&row.StartMonth,
				&row.EndDay,
				&row.EndMonth,
				&row.Year,
				&row.SplitYear)
			if err != nil {
				log.Println(err)
			}
			results = append(results, row)

		}
		err = rows.Err()
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		response, _ := json.Marshal(results)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(response))
	}
}

func (s *Server) ListCausesHandler() http.HandlerFunc {

	query := `
	SELECT DISTINCT
		death,
		death_id,
		description
	FROM 
		bom.causes_of_death
	ORDER BY 
		death ASC
	`

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]Causes, 0)
		var row Causes

		rows, err := s.DB.Query(context.TODO(), query)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Name, &row.ID, &row.Description)
			if err != nil {
				log.Println(err)
			}
			results = append(results, row)
		}
		err = rows.Err()
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		response, _ := json.Marshal(results)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, string(response))
	}

}
