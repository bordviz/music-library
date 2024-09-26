package dto

import (
	"fmt"
	"time"
)

type Filters struct {
	Group             any `json:"group"`
	Song              any `json:"song"`
	Text              any `json:"text"`
	ReleaseDateBefore any `json:"release_date_before"`
	ReleaseDateAfter  any `json:"release_date_after"`
}

func (f *Filters) Validate() error {
	if f.Group != nil {
		val, ok := f.Group.(string)
		if !ok {
			return fmt.Errorf("validation error: group filter must be a string")
		}
		f.Group = val
	}

	if f.Song != nil {
		val, ok := f.Song.(string)
		if !ok {
			return fmt.Errorf("validation error: song filter must be a string")
		}
		f.Song = val
	}

	if f.Text != nil {
		val, ok := f.Text.(string)
		if !ok {
			return fmt.Errorf("validation error: text filter must be a string")
		}
		f.Text = val
	}

	if f.ReleaseDateBefore != nil {
		val, ok := f.ReleaseDateBefore.(string)
		if !ok {
			return fmt.Errorf("validation error: release_date_before filter must be a string")
		}
		date, err := time.Parse("02.01.2006", val)
		if err != nil {
			return fmt.Errorf("invalid release_date_before format: %s, right format '16.09.2021'", val)
		}
		f.ReleaseDateBefore = date
	}

	if f.ReleaseDateAfter != nil {
		val, ok := f.ReleaseDateAfter.(string)
		if !ok {
			return fmt.Errorf("validation error: release_date_after filter must be a string")
		}
		date, err := time.Parse("02.01.2006", val)
		if err != nil {
			return fmt.Errorf("invalid release_date_after format: %s, right format '16.09.2021'", val)
		}
		f.ReleaseDateAfter = date
	}

	return nil
}
