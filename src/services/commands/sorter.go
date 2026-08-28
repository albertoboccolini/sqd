package commands

import (
	"cmp"
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

			switch order.Column {
			case models.NAME:
				compareResult = cmp.Compare(results[i].filePath, results[j].filePath)
			case models.CONTENT:
				compareResult = cmp.Compare(results[i].lineContent, results[j].lineContent)
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
