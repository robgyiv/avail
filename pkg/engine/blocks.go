package engine

import (
	"sort"
	"time"

	"github.com/robgyiv/avail/pkg/availability"
)

// GroupBlocksByDay groups time blocks by their calendar date in location.
//
// Blocks within a single day can carry different locations: work-hour edges come
// from the configured timezone, while an edge clipped by an event inherits that
// event's zone. Days are therefore keyed by the date in location, not by comparing
// time.Time values with == (which compares the *Location pointer as well, so two
// separately loaded copies of the same zone would split one day in two).
func GroupBlocksByDay(blocks []availability.TimeBlock, location *time.Location) []availability.Availability {
	if len(blocks) == 0 {
		return nil
	}
	if location == nil {
		location = time.UTC
	}

	// Sort blocks by start time
	sortedBlocks := make([]availability.TimeBlock, len(blocks))
	copy(sortedBlocks, blocks)
	sort.Slice(sortedBlocks, func(i, j int) bool {
		return sortedBlocks[i].Start.Before(sortedBlocks[j].Start)
	})

	// Group by day
	var result []availability.Availability
	var currentDate time.Time
	var currentBlocks []availability.TimeBlock

	for _, block := range sortedBlocks {
		blockDate := StartOfDay(block.Start.In(location))

		if !blockDate.Equal(currentDate) {
			// Save previous day's blocks if any
			if len(currentBlocks) > 0 {
				result = append(result, availability.Availability{
					Date:   currentDate,
					Blocks: currentBlocks,
				})
			}

			// Start new day
			currentDate = blockDate
			currentBlocks = []availability.TimeBlock{block}
		} else {
			currentBlocks = append(currentBlocks, block)
		}
	}

	// Don't forget the last day
	if len(currentBlocks) > 0 {
		result = append(result, availability.Availability{
			Date:   currentDate,
			Blocks: currentBlocks,
		})
	}

	return result
}
