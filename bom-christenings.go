package apiary

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// ChristeningsByYear describes a christening's description, total count, week number,
// week ID, and year.
type ChristeningsByYear struct {
	ChristeningsDesc string    `json:"christenings_desc"`
	TotalCount       NullInt64 `json:"count"`
	WeekNo           int       `json:"week_no"`
	WeekID           string    `json:"week_id"`
	Year             int       `json:"year"`
}

// Christenings describes a christening.
type Christenings struct {
	Name string `json:"name"`
}

// ChristeningsHandler returns the christenings for a given range of years. It expects a start year and
// end year as query parameters.
func (s *Server) ChristeningsHandler() http.HandlerFunc {

	query := `
	SELECT
		c.christening_desc,
		c.count,
		w.week_no,
		c.week_id,
		y.year
	FROM
		bom.christenings c
	JOIN
		bom.year y ON y.year_id = c.year_id
	JOIN
		bom.week w ON w.week_id = c.week_id
	WHERE
		year >= $1
		AND year < $2
	ORDER BY
		count
	LIMIT $3
	OFFSET $4;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		startYear := r.URL.Query().Get("start-year")
		endYear := r.URL.Query().Get("end-year")
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

		results := make([]ChristeningsByYear, 0)
		var row ChristeningsByYear

		rows, err := s.DB.Query(context.TODO(), query, startYearInt, endYearInt, limitInt, offsetInt)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(
				&row.ChristeningsDesc,
				&row.TotalCount,
				&row.WeekNo,
				&row.WeekID,
				&row.Year)
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

func (s *Server) ListChristeningsHandler() http.HandlerFunc {

	query := `
	SELECT DISTINCT
		christening_desc
	FROM 
		bom.christenings
	ORDER BY 
		christening_desc ASC
	`

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]Christenings, 0)
		var row Christenings

		rows, err := s.DB.Query(context.TODO(), query)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Name)
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
