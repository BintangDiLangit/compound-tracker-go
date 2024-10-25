package api

import (
	"database/sql"
	"fmt"
	"net/http"
)

func GetPointsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		address := r.URL.Query().Get("address")
		var points int
		err := db.QueryRow("SELECT SUM(points) FROM user_points WHERE address = $1", address).Scan(&points)
		if err != nil {
			http.Error(w, "Failed to retrieve points", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Address: %s, Points: %d", address, points)
	}
}
