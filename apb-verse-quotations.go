package dataapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// VerseQuotation is a single instance of a quotation
type VerseQuotation struct {
	Reference   string  `json:"reference"`
	Version     string  `json:"version"`
	DocID       string  `json:"document"`
	Date        string  `json:"date"`
	Probability float32 `json:"probability"`
	Title       string  `json:"title"`
}

// VerseQuotationsHandler returns the instances of quotations for a verse
func (s *Server) VerseQuotationsHandler() http.HandlerFunc {

	query := `
	SELECT q.reference_id, q.version, q.doc_id, q.date::text, q.probability,
	 	n.title_clean
	FROM apb.quotations q
	LEFT JOIN chronam.pages p ON q.doc_id = p.doc_id
	LEFT JOIN chronam.newspapers n ON p.lccn = n.lccn
	WHERE reference_id = $1 AND corpus = 'chronam'
	ORDER BY date;
	`
	stmt, err := s.Database.Prepare(query)
	if err != nil {
		log.Fatalln(err)
	}
	s.Statements["apb-verse-quotations"] = stmt // Will be closed at shutdown

	return func(w http.ResponseWriter, r *http.Request) {

		refs := r.URL.Query()["ref"]

		results := make([]VerseQuotation, 0)
		var row VerseQuotation

		rows, err := stmt.Query(refs[0])
		if err != nil {
			log.Println(err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&row.Reference, &row.Version, &row.DocID, &row.Date, &row.Probability, &row.Title)
			if err != nil {
				log.Println(err)
			}
			results = append(results, row)
		}
		err = rows.Err()
		if err != nil {
			log.Println(err)
		}

		if len(results) == 0 {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 Not found."))
		}

		response, _ := json.Marshal(results)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, string(response))
	}

}
