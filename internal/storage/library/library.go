package library

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"music-library/internal/domain/dto"
	"music-library/internal/domain/models"
	"music-library/internal/lib/logger/sl"
	"music-library/internal/lib/logger/with"
	"music-library/internal/lib/storage/query"
	"music-library/internal/lib/storage/tools"

	"github.com/jackc/pgx/v5"
)

type LibraryDB struct {
	log *slog.Logger
}

func NewLibraryDB(log *slog.Logger) *LibraryDB {
	return &LibraryDB{log: log}
}

func (db *LibraryDB) SaveSong(ctx context.Context, tx pgx.Tx, model dto.SongDB, requestID string) (int, error) {
	const op = "storage.library.SaveSong"

	db.log = with.WithOpAndRequestID(db.log, op, requestID)

	q := `
		INSERT INTO library 
		(group_name, song, release_date, text, patronymic)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`
	db.log.Debug("save new song query", slog.String("query", query.QueryToString(q)))

	var id int
	if err := tx.QueryRow(ctx, q, model.Group, model.Song, model.ReleaseDate, model.Text, model.Patronymic).Scan(&id); err != nil {
		db.log.Error("failed to save a new song", sl.Err(err))
		return 0, err
	}

	db.log.Info("new song was successfully saved", slog.Int("id", id))
	return id, nil
}

func (db *LibraryDB) GetLibray(ctx context.Context, tx pgx.Tx, filters dto.Filters, limit int, offset int, requestID string) ([]models.Song, error) {
	const op = "storage.library.GetLibrary"

	db.log = with.WithOpAndRequestID(db.log, op, requestID)

	filterStr, params, err := tools.GetFilters(filters)
	if err != nil {
		db.log.Error("failed to convert filters to SQL query", sl.Err(err))
		return nil, err
	}

	q := fmt.Sprintf(`
		SELECT id, group_name, song, to_char(release_date, 'DD.MM.YYYY'), text, patronymic
		FROM library
		WHERE %s
		LIMIT $%d
		OFFSET $%d;
	`, filterStr, len(params)+1, len(params)+2)

	db.log.Debug("get library query", slog.String("query", query.QueryToString(q)))

	params = append(params, limit, offset)

	rows, err := tx.Query(ctx, q, params...)
	if err != nil {
		db.log.Error("failed to get library", sl.Err(err))
		return nil, err
	}
	defer rows.Close()

	var songs []models.Song
	for rows.Next() {
		var song models.Song
		if err := rows.Scan(&song.ID, &song.Group, &song.Song, &song.ReleaseDate, &song.Text, &song.Patronymic); err != nil {
			db.log.Error("failed to scan row", sl.Err(err))
			return nil, err
		}
		songs = append(songs, song)
	}

	if rows.Err() != nil {
		db.log.Error("failed to scan rows", sl.Err(rows.Err()))
		return nil, err
	}

	db.log.Info("library was successfully retrieved", slog.Int("count", len(songs)))
	return songs, nil
}

func (db *LibraryDB) GetSongText(ctx context.Context, tx pgx.Tx, songID int, requestID string) (string, error) {
	const op = "storage.library.GetSongText"

	db.log = with.WithOpAndRequestID(db.log, op, requestID)

	q := `
        SELECT text
        FROM library
        WHERE id = $1
    `
	db.log.Debug("get song text query", slog.String("query", query.QueryToString(q)))

	var text string
	if err := tx.QueryRow(ctx, q, songID).Scan(&text); err != nil {
		if err == pgx.ErrNoRows {
			db.log.Error("song text not found", slog.Int("song_id", songID))
			return "", errors.New("song not found")
		}
		db.log.Error("failed to get song text", sl.Err(err))
	}

	db.log.Info("song text was successfully retrieved", slog.Int("song_id", songID))
	return text, nil
}

func (db *LibraryDB) DeleteSong(ctx context.Context, tx pgx.Tx, songID int, requestID string) error {
	const op = "storage.library.DeleteSong"

	db.log = with.WithOpAndRequestID(db.log, op, requestID)

	q := `
        DELETE FROM library
        WHERE id = $1
		RETURNING id;
    `
	db.log.Debug("delete song query", slog.String("query", query.QueryToString(q)))

	var id int
	if err := tx.QueryRow(ctx, q, songID).Scan(&id); err != nil {
		if err == pgx.ErrNoRows {
			db.log.Error("song not found", slog.Int("song_id", songID))
			return errors.New("song not found")
		}
		db.log.Error("failed to delete song", sl.Err(err))
		return err
	}

	if id == 0 {
		db.log.Error("failed to delete song", slog.Int("song_id", songID))
		return errors.New("failed to delete song")
	}

	db.log.Info("song was successfully deleted", slog.Int("id", id))
	return nil
}

func (db *LibraryDB) UpdateSong(ctx context.Context, tx pgx.Tx, updateModel dto.UpdateSong, requestID string) error {
	const op = "storage.library.UpdateSong"

	db.log = with.WithOpAndRequestID(db.log, op, requestID)
	strParams, params := tools.GetUpdateParams(updateModel)

	q := fmt.Sprintf(`
		UPDATE library
		SET %s
		WHERE id = $%d
		RETURNING id;
	`, strParams, len(params)+1)

	params = append(params, updateModel.ID)

	db.log.Debug("update song query", slog.String("query", query.QueryToString(q)))

	var id int
	if err := tx.QueryRow(ctx, q, params...).Scan(&id); err != nil {
		if err == pgx.ErrNoRows {
			db.log.Error("song not found", slog.Int("song_id", updateModel.ID))
			return errors.New("song not found")
		}
		db.log.Error("failed to update song", sl.Err(err))
		return err
	}

	if id == 0 {
		db.log.Error("failed to update song", slog.Int("song_id", updateModel.ID))
		return errors.New("failed to update song")
	}

	db.log.Info("song was successfully updated", slog.Int("id", id))
	return nil
}
