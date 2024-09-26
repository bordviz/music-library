package dto

import (
	"fmt"
	"music-library/internal/lib/validator"
	"strings"
	"time"
)

type SongRequest struct {
	Group string `json:"group" validate:"required" example:"Muse"`
	Song  string `json:"song" validate:"required" example:"Supermassive Black Hole"`
}

func (r *SongRequest) Validate() error {
	r.Group = strings.TrimSpace(r.Group)
	r.Song = strings.TrimSpace(r.Song)

	if err := validator.Validate(r); err != "" {
		return fmt.Errorf("validation error: %s", err)
	}
	return nil
}

type SongDB struct {
	Group       string    `json:"group"`
	Song        string    `json:"song"`
	ReleaseDate time.Time `json:"releaseDate"`
	Text        string    `json:"text"`
	Patronymic  string    `json:"patronymic"`
}

type Song struct {
	Group       string `json:"group" validate:"required"`
	Song        string `json:"song" validate:"required"`
	ReleaseDate string `json:"releaseDate" validate:"required"`
	Text        string `json:"text" validate:"required"`
	Patronymic  string `json:"patronymic" validate:"required"`
}

func (s *Song) Validate() error {
	s.Group = strings.TrimSpace(s.Group)
	s.Song = strings.TrimSpace(s.Song)
	s.ReleaseDate = strings.TrimSpace(s.ReleaseDate)
	s.Text = strings.TrimSpace(s.Text)
	s.Patronymic = strings.TrimSpace(s.Patronymic)

	if err := validator.Validate(s); err != "" {
		return fmt.Errorf("validation error: %s", err)
	}
	return nil
}

func (s *Song) ToDBModel() (SongDB, error) {
	releaseDate, err := time.Parse("02.01.2006", s.ReleaseDate)
	if err != nil {
		return SongDB{}, fmt.Errorf("invalid release_date format: %s, right format '16.09.2021'", s.ReleaseDate)
	}

	return SongDB{
		Group:       s.Group,
		Song:        s.Song,
		ReleaseDate: releaseDate,
		Text:        s.Text,
		Patronymic:  s.Patronymic,
	}, nil
}

type UpdateSong struct {
	ID          int `json:"id" validate:"required"`
	Group       any `json:"group"`
	Song        any `json:"song"`
	ReleaseDate any `json:"releaseDate"`
	Text        any `json:"text"`
	Patronymic  any `json:"patronymic"`
}

func (u *UpdateSong) Validate() error {

	if err := validator.Validate(u); err != "" {
		return fmt.Errorf("validation error: %s", err)
	}

	if u.Group != nil {
		val, ok := u.Group.(string)
		if !ok {
			return fmt.Errorf("validation error: group filter must be a string")
		}
		u.Group = val
	}

	if u.Song != nil {
		val, ok := u.Song.(string)
		if !ok {
			return fmt.Errorf("validation error: song filter must be a string")
		}
		u.Song = val
	}

	if u.Text != nil {
		val, ok := u.Text.(string)
		if !ok {
			return fmt.Errorf("validation error: text filter must be a string")
		}
		u.Text = val
	}

	if u.ReleaseDate != nil {
		val, ok := u.ReleaseDate.(string)
		if !ok {
			return fmt.Errorf("validation error: release_date filter must be a string")
		}
		date, err := time.Parse("02.01.2006", val)
		if err != nil {
			return fmt.Errorf("invalid release_date_before format: %s, right format '16.09.2021'", val)
		}
		u.ReleaseDate = date
	}

	if u.Patronymic != nil {
		val, ok := u.Patronymic.(string)
		if !ok {
			return fmt.Errorf("validation error: patronymic filter must be a string")
		}
		u.Patronymic = val
	}

	return nil
}
