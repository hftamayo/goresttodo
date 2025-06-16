package task

import (
    "testing"

    "github.com/hftamayo/gotodo/pkg/utils"
    "github.com/stretchr/testify/assert"
)

func TestValidatePaginationQuery(t *testing.T) {
    tests := []struct {
        name     string
        input    CursorPaginationQuery
        expected CursorPaginationQuery
    }{
        {
            name: "valid query with all parameters",
            input: CursorPaginationQuery{
                Cursor: "cursor123",
                Limit:  10,
                Order:  "asc",
            },
            expected: CursorPaginationQuery{
                Cursor: "cursor123",
                Limit:  10,
                Order:  "asc",
            },
        },
        {
            name: "zero limit should use default",
            input: CursorPaginationQuery{
                Cursor: "cursor123",
                Limit:  0,
                Order:  "desc",
            },
            expected: CursorPaginationQuery{
                Cursor: "cursor123",
                Limit:  utils.DefaultLimit,
                Order:  "desc",
            },
        },
        {
            name: "negative limit should use default",
            input: CursorPaginationQuery{
                Cursor: "cursor123",
                Limit:  -5,
                Order:  "desc",
            },
            expected: CursorPaginationQuery{
                Cursor: "cursor123",
                Limit:  utils.DefaultLimit,
                Order:  "desc",
            },
        },
        {
            name: "limit exceeding max should be capped",
            input: CursorPaginationQuery{
                Cursor: "cursor123",
                Limit:  utils.MaxLimit + 10,
                Order:  "desc",
            },
            expected: CursorPaginationQuery{
                Cursor: "cursor123",
                Limit:  utils.MaxLimit,
                Order:  "desc",
            },
        },
        {
            name: "invalid order should use default",
            input: CursorPaginationQuery{
                Cursor: "cursor123",
                Limit:  10,
                Order:  "invalid",
            },
            expected: CursorPaginationQuery{
                Cursor: "cursor123",
                Limit:  10,
                Order:  utils.DefaultOrder,
            },
        },
        {
            name: "order should be normalized to lowercase",
            input: CursorPaginationQuery{
                Cursor: "cursor123",
                Limit:  10,
                Order:  "ASC",
            },
            expected: CursorPaginationQuery{
                Cursor: "cursor123",
                Limit:  10,
                Order:  "asc",
            },
        },
        {
            name: "cursor should be trimmed",
            input: CursorPaginationQuery{
                Cursor: "  cursor123  ",
                Limit:  10,
                Order:  "desc",
            },
            expected: CursorPaginationQuery{
                Cursor: "cursor123",
                Limit:  10,
                Order:  "desc",
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := validatePaginationQuery(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestValidatePagePaginationQuery(t *testing.T) {
    tests := []struct {
        name     string
        input    PagePaginationQuery
        expected PagePaginationQuery
    }{
        {
            name: "valid query with all parameters",
            input: PagePaginationQuery{
                Page:  2,
                Limit: 10,
                Order: "asc",
            },
            expected: PagePaginationQuery{
                Page:  2,
                Limit: 10,
                Order: "asc",
            },
        },
        {
            name: "zero limit should use default",
            input: PagePaginationQuery{
                Page:  1,
                Limit: 0,
                Order: "desc",
            },
            expected: PagePaginationQuery{
                Page:  1,
                Limit: utils.DefaultLimit,
                Order: "desc",
            },
        },
        {
            name: "negative limit should use default",
            input: PagePaginationQuery{
                Page:  1,
                Limit: -5,
                Order: "desc",
            },
            expected: PagePaginationQuery{
                Page:  1,
                Limit: utils.DefaultLimit,
                Order: "desc",
            },
        },
        {
            name: "limit exceeding max should be capped",
            input: PagePaginationQuery{
                Page:  1,
                Limit: utils.MaxLimit + 10,
                Order: "desc",
            },
            expected: PagePaginationQuery{
                Page:  1,
                Limit: utils.MaxLimit,
                Order: "desc",
            },
        },
        {
            name: "zero page should be set to 1",
            input: PagePaginationQuery{
                Page:  0,
                Limit: 10,
                Order: "desc",
            },
            expected: PagePaginationQuery{
                Page:  1,
                Limit: 10,
                Order: "desc",
            },
        },
        {
            name: "negative page should be set to 1",
            input: PagePaginationQuery{
                Page:  -5,
                Limit: 10,
                Order: "desc",
            },
            expected: PagePaginationQuery{
                Page:  1,
                Limit: 10,
                Order: "desc",
            },
        },
        {
            name: "page exceeding max should be capped",
            input: PagePaginationQuery{
                Page:  200,
                Limit: 10,
                Order: "desc",
            },
            expected: PagePaginationQuery{
                Page:  100,
                Limit: 10,
                Order: "desc",
            },
        },
        {
            name: "invalid order should use default",
            input: PagePaginationQuery{
                Page:  1,
                Limit: 10,
                Order: "invalid",
            },
            expected: PagePaginationQuery{
                Page:  1,
                Limit: 10,
                Order: utils.DefaultOrder,
            },
        },
        {
            name: "order should be normalized to lowercase",
            input: PagePaginationQuery{
                Page:  1,
                Limit: 10,
                Order: "ASC",
            },
            expected: PagePaginationQuery{
                Page:  1,
                Limit: 10,
                Order: "asc",
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := validatePagePaginationQuery(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
} 