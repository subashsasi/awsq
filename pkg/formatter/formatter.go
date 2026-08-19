package formatter

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Print outputs data in the specified format (table, json, csv)
func Print(format string, headers []string, rows [][]string) {
	switch format {
	case "json":
		printJSON(headers, rows)
	case "csv":
		printCSV(headers, rows)
	default:
		printTable(headers, rows)
	}
}

func printTable(headers []string, rows [][]string) {
	if len(rows) == 0 {
		fmt.Println("No resources found.")
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, col := range row {
			if i < len(widths) && len(col) > widths[i] {
				widths[i] = len(col)
			}
		}
	}

	// Cap max width at 40 characters
	for i := range widths {
		if widths[i] > 40 {
			widths[i] = 40
		}
	}

	// Print header
	for i, h := range headers {
		fmt.Printf("%-*s  ", widths[i], h)
	}
	fmt.Println()

	// Print separator
	for i := range headers {
		fmt.Printf("%s  ", strings.Repeat("-", widths[i]))
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		for i, col := range row {
			if i < len(widths) {
				if len(col) > widths[i] {
					col = col[:widths[i]-1] + "…"
				}
				fmt.Printf("%-*s  ", widths[i], col)
			}
		}
		fmt.Println()
	}
}

func printJSON(headers []string, rows [][]string) {
	var result []map[string]string
	for _, row := range rows {
		item := make(map[string]string)
		for i, h := range headers {
			if i < len(row) {
				item[strings.ToLower(h)] = row[i]
			}
		}
		result = append(result, item)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
}

func printCSV(headers []string, rows [][]string) {
	w := csv.NewWriter(os.Stdout)
	w.Write(headers)
	for _, row := range rows {
		w.Write(row)
	}
	w.Flush()
}
