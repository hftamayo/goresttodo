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

    if query.Cursor != "" {
        query.Cursor = strings.TrimSpace(query.Cursor)
    }    

    return query
}

func validatePagePaginationQuery(query PagePaginationQuery) PagePaginationQuery {
    // Validate and set limit
    if query.Limit <= 0 {
        query.Limit = utils.DefaultLimit
    }
    if query.Limit > utils.MaxLimit {
        query.Limit = utils.MaxLimit
    }
    
    // Ensure page is at least 1
    if query.Page <= 0 {
        query.Page = 1
    }
    
    // Normalize order to lowercase and set default
    query.Order = strings.ToLower(query.Order)
    if query.Order != "asc" && query.Order != "desc" {
        query.Order = utils.DefaultOrder
    }
    
    return query
}