package utils

import "math"

func Paginate(total int64, page, perPage int) (map[string]interface{}, error) {
	// Avoid division by zero
	if perPage <= 0 {
		perPage = 1
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))

	// Determine next and previous pages
	var nextPage, prevPage *int
	if page < totalPages {
		next := page + 1
		nextPage = &next
	}
	if page > 1 {
		prev := page - 1
		prevPage = &prev
	}

	// Construct and return pagination data
	return map[string]interface{}{
		"current_page":  page,
		"item_per_page": perPage,
		"next_page":     nextPage,
		"previous_page": prevPage,
		"total_count":   total,
		"total_pages":   totalPages,
	}, nil
}
