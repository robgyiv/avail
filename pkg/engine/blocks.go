package engine

import (
	"sort"
	"time"

	"github.com/robgyiv/availability/pkg/availability"
)

// GroupBlocksByDay groups time blocks by their date.
func GroupBlocksByDay(blocks []availability.TimeBlock) []availability.Availability {
	if len(blocks) == 0 {
		return nil
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
		blockDate := block.Start.Truncate(24 * time.Hour)

		if blockDate != currentDate {
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
