package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/covista/commons/proto"
)

// given a request for downloading DiagnosisKeys, build the SQL query that returns
// the keys that match the filter. We know because of 'checkGetKeyRequest' that
// there is at least one filter defined in 'request'
func buildQuery(request *proto.GetKeyRequest) (string, []interface{}, error) {
	var query_values []interface{}
	var query = `SELECT TEK, ENIN FROM reported_keys WHERE `
	var clauses []string
	var suffix string
	var err error

	// if health authority identifier is provided...
	if len(request.HAK) > 0 {
		query_values = append(query_values, request.HAK)
		clauses = append(clauses, fmt.Sprintf("authority_id = $%d", len(query_values)))
		suffix += `JOIN authorization_keys USING (authorization_key)
			       JOIN health_authorities USING (api_key)`
	}

	// if an ENIN is provided, round to the nearest day and default to [ENIN, ENIN + 1 day]
	if request.ENIN > 0 {
		start := eninToTimestamp(request.ENIN)
		end := start.Add(24 * time.Hour)
		query_values = append(query_values, start)
		clauses = append(clauses, fmt.Sprintf("enin >= $%d", len(query_values)))
		query_values = append(query_values, end)
		clauses = append(clauses, fmt.Sprintf("enin <= $%d", len(query_values)))
	}

	// if historical range [start, end] dates are provided, use that range.
	// if historical range [days] is provided, generate filter for the last N days
	// starting at [start_date]
	if len(request.Hrange.StartDate) > 0 || request.Hrange.Days > 0 {
		// default to current date if start_date not defined
		var start, end time.Time
		if len(request.Hrange.StartDate) == 0 {
			end = time.Now()
		} else {
			end, err = time.Parse(time.RFC3339, request.Hrange.StartDate)
			if err != nil {
				return query, query_values, err
			}
		}
		end = end.UTC().Truncate(24 * time.Hour)
		num_days := max(request.Hrange.Days, 1)
		start = end.Add(time.Duration(num_days) * -24 * time.Hour)
		query_values = append(query_values, start)
		clauses = append(clauses, fmt.Sprintf("enin >= $%d", len(query_values)))
		query_values = append(query_values, end)
		clauses = append(clauses, fmt.Sprintf("enin <= $%d", len(query_values)))
	}

	query = fmt.Sprintf("%s %s %s", query, strings.Join(clauses, " AND "), suffix)
	return query, query_values, nil
}
