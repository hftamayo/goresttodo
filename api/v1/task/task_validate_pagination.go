package task

import (
	"strings"
)

func validatePaginationQuery(query CursorPaginationQuery) CursorPaginationQuery {
    if query.Limit <= 0 {
        query.Limit = DefaultLimit
    }
    if query.Limit > MaxLimit {
        query.Limit = MaxLimit
    }
    
    query.Order = strings.ToLower(query.Order)
    if query.Order != "asc" && query.Order != "desc" {
        query.Order = DefaultOrder
    }

    return query
}