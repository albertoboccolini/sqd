package commands

import (
	"sort"

	"github.com/overthinkinglabs/sqd/src/models"
)

type Sorter struct{}

func NewSorter() *Sorter {
	return &Sorter{}
}

func (sorter *Sorter) sortResults(results []searchResult, orderBy []models.OrderBy) {
	if len(orderBy) == 0 {
		return
	}

	sort.Slice(results, func(i, j int) bool {
		for _, order := range orderBy {
			var compareResult int

			if order.Column == models.NAME {
				if results[i].filePath < results[j].filePath {
					compareResult = -1
				}

				if results[i].filePath > results[j].filePath {
					compareResult = 1
				}
			}

			if order.Column == models.CONTENT {
				if results[i].lineContent < results[j].lineContent {
					compareResult = -1
				}

				if results[i].lineContent > results[j].lineContent {
					compareResult = 1
				}
			}

			if compareResult != 0 {
				if order.Direction == models.DESC {
					return compareResult > 0
				}

				return compareResult < 0
			}
		}

		return false
	})
}
