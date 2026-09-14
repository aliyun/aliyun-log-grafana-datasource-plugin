package main

import (
	"fmt"
	"regexp"
	"strings"
)

// InterpolateMacros replaces Grafana macros with SLS SQL equivalents
// Supported macros:
// $__time(col) -> to_unixtime(col) as time
// $__timeFilter(col) -> col >= from AND col < to
// $__timeGroup(col, 'interval', [fill]) -> time_series(col, 'interval', '%Y-%m-%d %H:%i:%s', fill)
// $__timeGroupAlias(col, 'interval') -> time_series(...) as time
func InterpolateMacros(query string, from, to int64) string {
	// $__time(dateColumn)
	// Example: $__time(log_time) -> to_unixtime(log_time) as time
	timeReg := regexp.MustCompile(`\$__time\(([^)]+)\)`)
	query = timeReg.ReplaceAllString(query, "to_unixtime($1) as time")

	// $__timeFilter(dateColumn)
	// Example: $__timeFilter(__time__) -> __time__ >= 1600000000 AND __time__ < 1600003600
	timeFilterReg := regexp.MustCompile(`\$__timeFilter\(([^)]+)\)`)
	query = timeFilterReg.ReplaceAllStringFunc(query, func(match string) string {
		parts := timeFilterReg.FindStringSubmatch(match)
		if len(parts) == 2 {
			col := parts[1]
			return fmt.Sprintf("%s >= %d AND %s < %d", col, from, col, to)
		}
		return match
	})

	// $__timeGroup(dateColumn, '5m', fill)
	// Matches: $__timeGroup(col, 'interval' [, fill])
	// Note: interval is expected to be quoted, fill is optional
	timeGroupReg := regexp.MustCompile(`\$__timeGroup\(\s*([^,]+)\s*,\s*'([^']+)'(?:\s*,\s*([^)]+))?\s*\)`)
	query = timeGroupReg.ReplaceAllStringFunc(query, func(match string) string {
		parts := timeGroupReg.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		col := parts[1]
		interval := parts[2]
		fill := "0" // Default fill
		
		if len(parts) > 3 && parts[3] != "" {
			fill = strings.TrimSpace(parts[3])
		}

		// Map fill strategies
		// MySQL: 0, NULL, previous
		// SLS: '0', 'null', 'last'
		slsFill := "'0'"
		switch strings.ToLower(fill) {
		case "null":
			slsFill = "'null'"
		case "previous":
			slsFill = "'last'"
		case "0":
			slsFill = "'0'"
		default:
			// Treat as explicit value, wrap in quotes for SLS if it looks like a number or string
            // SLS time_series padding expects a string literal '...'
			slsFill = fmt.Sprintf("'%s'", fill)
		}

		return fmt.Sprintf("time_series(%s, '%s', '%%Y-%%m-%%d %%H:%%i:%%s', %s)", col, interval, slsFill)
	})

	// $__timeGroupAlias(dateColumn, '5m')
	// Similar to timeGroup but adds 'as time'
	timeGroupAliasReg := regexp.MustCompile(`\$__timeGroupAlias\(\s*([^,]+)\s*,\s*'([^']+)'\s*\)`)
	query = timeGroupAliasReg.ReplaceAllStringFunc(query, func(match string) string {
		parts := timeGroupAliasReg.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		col := parts[1]
		interval := parts[2]
		
		return fmt.Sprintf("time_series(%s, '%s', '%%Y-%%m-%%d %%H:%%i:%%s', '0') as time", col, interval)
	})

	return query
}
