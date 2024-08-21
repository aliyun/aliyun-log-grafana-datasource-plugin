package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetQueryCount(t *testing.T) {
	queryInfo := &QueryInfo{
		QueryType:    "type",
		QueryMode:    "mode",
		Query:        "*",
		Xcol:         "xcol",
		Ycol:         "ycol",
		LogsPerPage:  10,
		CurrentPage:  1,
		TotalResults: 101,
	}
	count := GetQueryCount(queryInfo)
	assert.True(t, count == 11)

	queryInfo.Query = "*|select count(1) as count"
	count = GetQueryCount(queryInfo)
	assert.True(t, count == 1)

	queryInfo.TotalResults = 1000
	queryInfo.Query = "*"
	count = GetQueryCount(queryInfo)
	assert.True(t, count == 100)

	queryInfo.TotalResults = 5000
	queryInfo.Query = "*"
	count = GetQueryCount(queryInfo)
	assert.True(t, count == 500)

	queryInfo.TotalResults = 6000
	queryInfo.Query = "*"
	count = GetQueryCount(queryInfo)
	assert.True(t, count == 500)

	queryInfo.TotalResults = 5
	queryInfo.Query = "*"
	count = GetQueryCount(queryInfo)
	assert.True(t, count == 1)
	assert.True(t, queryInfo.LogsPerPage == 5)
}
