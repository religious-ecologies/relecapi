package apiary

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Parish describes a parish name, canonical name, and unique ID.
type Parish struct {
	ParishID      int    `json:"id"`
	Name          string `json:"name"`
	CanonicalName string `json:"canonical_name"`
}

// ParishesHandler returns a list of unique parish IDs and names.
func (s *Server) ParishesHandler() http.HandlerFunc {

	query := `
	SELECT id, parish_name, canonical_name 
	FROM bom.parishes
	ORDER BY canonical_name;
	`

	return func(w http.ResponseWriter, r *http.Request) {
		results := make([]Parish, 0)
		var row Parish

		rows, err := s.DB.Query(context.TODO(), query)
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.ParishID, &row.Name, &row.CanonicalName)
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
