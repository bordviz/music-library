package library

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"music-library/internal/config"
	"music-library/internal/domain/dto"
	"music-library/internal/domain/models"
	"music-library/internal/lib/logger/sl"
	"music-library/internal/lib/logger/with"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LibraryService struct {
	log  *slog.Logger
	pool *pgxpool.Pool
	db   LibraryDB
	cfg  config.LibraryServer
}

type LibraryDB interface {
	SaveSong(ctx context.Context, tx pgx.Tx, model dto.SongDB, requestID string) (int, error)
	GetLibray(ctx context.Context, tx pgx.Tx, filters dto.Filters, limit int, offset int, requestID string) ([]models.Song, error)
	GetSongText(ctx context.Context, tx pgx.Tx, songID int, requestID string) (string, error)
	DeleteSong(ctx context.Context, tx pgx.Tx, songID int, requestID string) error
	UpdateSong(ctx context.Context, tx pgx.Tx, updateModel dto.UpdateSong, requestID string) error
}

func NewLibraryService(log *slog.Logger, pool *pgxpool.Pool, db LibraryDB, cfg config.LibraryServer) *LibraryService {
	return &LibraryService{log: log, pool: pool, db: db, cfg: cfg}
}

func (s *LibraryService) SaveSong(ctx context.Context, model dto.SongRequest, requestID string) (int, error) {
	const op = "library.service.SaveSong"

	s.log = with.WithOpAndRequestID(s.log, op, requestID)

	url := fmt.Sprintf("%s://%s:%d/info", s.cfg.Protocol, s.cfg.Host, s.cfg.Port)
	s.log.Debug("request url", slog.String("url", url))

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		s.log.Error("failed to get song info", sl.Err(err))
		return 0, err
	}

	q := req.URL.Query()
	q.Set("group", model.Group)
	q.Set("song", model.Song)
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.log.Error("failed to make request", sl.Err(err))
		return 0, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	fmt.Println(string(body))

	var song dto.Song
	err = json.Unmarshal(body, &song)
	if err != nil {
		s.log.Error("failed to unmarshal song info", sl.Err(err))
		return 0, err
	}

	if err := song.Validate(); err != nil {
		s.log.Error("validation error in song info", sl.Err(err))
		return 0, err
	}

	modelDB, err := song.ToDBModel()
	if err != nil {
		s.log.Error("failed to convert song to db model", sl.Err(err))
		return 0, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Error("failed to begin transaction", sl.Err(err))
		return 0, err
	}
	defer tx.Rollback(ctx)

	id, err := s.db.SaveSong(ctx, tx, modelDB, requestID)
	if err != nil {
		s.log.Error("failed to save song", sl.Err(err))
		return 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		s.log.Error("failed to commit transaction", sl.Err(err))
		return 0, err
	}

	s.log.Info("song was successfully saved", slog.Int("id", id))
	return id, nil
}

func (s *LibraryService) GetLibrary(ctx context.Context, filters dto.Filters, limit int, offset int, requestID string) ([]models.Song, error) {
	const op = "library.service.GetLibrary"

	s.log = with.WithOpAndRequestID(s.log, op, requestID)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Error("failed to begin transaction", sl.Err(err))
		return nil, err
	}
	defer tx.Rollback(ctx)

	songs, err := s.db.GetLibray(ctx, tx, filters, limit, offset, requestID)
	if err != nil {
		s.log.Error("failed to get library", sl.Err(err))
		return nil, err
	}

	s.log.Info("library successfully fetched", slog.Int("songs_count", len(songs)))
	return songs, nil
}

func (s *LibraryService) GetSongText(ctx context.Context, songID int, couplet int, requestID string) (string, error) {
	const op = "library.service.GetSongText"

	s.log = with.WithOpAndRequestID(s.log, op, requestID)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Error("failed to begin transaction", sl.Err(err))
		return "", err
	}
	defer tx.Rollback(ctx)

	text, err := s.db.GetSongText(ctx, tx, songID, requestID)
	if err != nil {
		s.log.Error("failed to get song text", sl.Err(err))
		return "", err
	}

	cup := []rune{}
	counter := 1
Loop:
	for _, symb := range text {
		if symb == '\n' && len(cup) != 0 && cup[len(cup)-1] == '\n' {
			if counter == couplet {
				break Loop
			}
			cup = cup[:0]
			counter++
			continue Loop
		} else {
			cup = append(cup, symb)
		}
	}

	s.log.Info("song text successfully fetched", slog.Int("song_id", songID), slog.Int("couplet", couplet))
	return string(cup), nil
}

func (s *LibraryService) DeleteSong(ctx context.Context, songID int, requestID string) error {
	const op = "library.service.DeleteSong"

	s.log = with.WithOpAndRequestID(s.log, op, requestID)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Error("failed to begin transaction", sl.Err(err))
		return err
	}
	defer tx.Rollback(ctx)

	err = s.db.DeleteSong(ctx, tx, songID, requestID)
	if err != nil {
		s.log.Error("failed to delete song", sl.Err(err))
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		s.log.Error("failed to commit transaction", sl.Err(err))
		return err
	}

	s.log.Info("song was successfully deleted")
	return nil
}

func (s *LibraryService) UpdateSong(ctx context.Context, updateModel dto.UpdateSong, requestID string) error {
	const op = "library.service.UpdateSong"

	s.log = with.WithOpAndRequestID(s.log, op, requestID)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.log.Error("failed to begin transaction", sl.Err(err))
		return err
	}
	defer tx.Rollback(ctx)

	err = s.db.UpdateSong(ctx, tx, updateModel, requestID)
	if err != nil {
		s.log.Error("failed to update song", sl.Err(err))
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		s.log.Error("failed to commit transaction", sl.Err(err))
		return err
	}

	s.log.Info("song was successfully updated")
	return nil
}
