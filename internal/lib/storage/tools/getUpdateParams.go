package tools

import (
	"fmt"
	"music-library/internal/domain/dto"
)

func GetUpdateParams(model dto.UpdateSong) (string, []any) {
	params := make([]any, 0, 6)
	var setStr string

	if model.Group != nil {
		if setStr != "" {
			setStr += ", "
		}
		params = append(params, model.Group)
		setStr += fmt.Sprintf("group_name = $%d", len(params))
	}

	if model.Song != nil {
		if setStr != "" {
			setStr += ", "
		}
		params = append(params, model.Song)
		setStr += fmt.Sprintf("song = $%d", len(params))
	}

	if model.Text != nil {
		if setStr != "" {
			setStr += ", "
		}
		params = append(params, model.Text)
		setStr += fmt.Sprintf("text = $%d", len(params))
	}

	if model.ReleaseDate != nil {
		if setStr != "" {
			setStr += ", "
		}
		params = append(params, model.ReleaseDate)
		setStr += fmt.Sprintf("release_date = $%d", len(params))
	}

	if model.Patronymic != nil {
		if setStr != "" {
			setStr += ", "
		}
		params = append(params, model.Patronymic)
		setStr += fmt.Sprintf("patronymic = $%d", len(params))
	}

	return setStr, params
}
