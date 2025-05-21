package task

import (
	"strings"

	"github.com/hftamayo/gotodo/pkg/utils"
)

func validatePaginationQuery(query CursorPaginationQuery) CursorPaginationQuery {
    if query.Limit <= 0 {
        query.Limit = utils.DefaultLimit
    }
    if query.Limit > utils.MaxLimit {
        query.Limit = utils.MaxLimit
    }
    
    query.Order = strings.ToLower(query.Order)
    if query.Order != "asc" && query.Order != "desc" {
        query.Order = utils.DefaultOrder
    }

    return query
}