package model

// New rules for models:
// 1. Methods: load, save, delete
// 2. Free functions: exists, all
// 3. All functions should take db and tx as parameters
// 4. Don't rollback in functions, let the caller handle it
// 5. All returns list of objects, not pointers to objects
// 6. Save is not atomic unless tx is provided. If row is updates, pk should not be changed
// 7. Load loads by primary key, returns true if found, false if not found

type ComparationOperation struct {
	FieldValue any
	Operation  string
}

// PrepareQuery prepares a SQL query with placeholders for the provided filter map.
// It's safe to use if provided query doesn't contain placeholders and WHERE clause already.
// Also, it does not work if WHERE clause syntaxically can't come after the provided query.
func PrepareQuery(query string, filter map[string]any) (string, []any) {
	if len(filter) == 0 {
		return query, nil
	}

	query += " WHERE "

	var queryParams []any
	for key, value := range filter {
		if len(queryParams) > 0 {
			query += " AND "
		}

		compOp, ok := value.(ComparationOperation)
		if ok {
			query += key + compOp.Operation + "?"
			queryParams = append(queryParams, compOp.FieldValue)
		} else {
			query += key + "=?"
			queryParams = append(queryParams, value)
		}
	}

	return query, queryParams
}
