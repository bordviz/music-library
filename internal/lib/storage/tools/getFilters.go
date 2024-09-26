package tools

import (
	"errors"
	"fmt"
	"music-library/internal/domain/dto"
	"strings"
)

func GetFilters(filters dto.Filters) (string, []any, error) {
	var filterStr string
	params := make([]any, 0, 5)

	typeErr := errors.New("failed to convert filters")

	toStr, ok := filters.Group.(string)
	if filters.Group != nil {
		if !ok {
			return "", nil, typeErr
		}
		if filterStr != "" {
			filterStr += " AND "
		}
		params = append(params, "%"+strings.ToLower(toStr)+"%")
		filterStr += fmt.Sprintf("LOWER(group_name) LIKE $%d", len(params))
	}

	if filters.Song != nil {
		toStr, ok := filters.Song.(string)
		if !ok {
			return "", nil, typeErr
		}
		if filterStr != "" {
			filterStr += " AND "
		}
		params = append(params, "%"+strings.ToLower(toStr)+"%")
		filterStr += fmt.Sprintf("LOWER(song) LIKE $%d", len(params))
	}

	if filters.Text != nil {
		toStr, ok := filters.Text.(string)
		if !ok {
			return "", nil, typeErr
		}
		if filterStr != "" {
			filterStr += " AND "
		}
		params = append(params, "%"+strings.ToLower(toStr)+"%")
		filterStr += fmt.Sprintf("LOWER(text) LIKE $%d", len(params))
	}

	if filters.ReleaseDateBefore != nil {
		if filterStr != "" {
			filterStr += " AND "
		}
		params = append(params, filters.ReleaseDateBefore)
		filterStr += fmt.Sprintf("release_date <= $%d", len(params))
	}

	if filters.ReleaseDateAfter != nil {
		if filterStr != "" {
			filterStr += " AND "
		}
		params = append(params, filters.ReleaseDateAfter)
		filterStr += fmt.Sprintf("release_date >= $%d", len(params))
	}

	return filterStr, params, nil
}
