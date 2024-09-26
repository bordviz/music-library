package library

import (
	"context"
	"log/slog"
	"music-library/internal/domain/dto"
	"music-library/internal/domain/models"
	"music-library/internal/handlers"
	"music-library/internal/lib/logger/sl"
	"music-library/internal/lib/logger/with"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Handler struct {
	log     *slog.Logger
	service LibraryService
}

type LibraryService interface {
	SaveSong(ctx context.Context, model dto.SongRequest, requestID string) (int, error)
	GetLibrary(ctx context.Context, filters dto.Filters, limit int, offset int, requestID string) ([]models.Song, error)
	GetSongText(ctx context.Context, songID int, couplet int, requestID string) (string, error)
	DeleteSong(ctx context.Context, songID int, requestID string) error
	UpdateSong(ctx context.Context, updateModel dto.UpdateSong, requestID string) error
}

func NewHandler(log *slog.Logger, service LibraryService) *Handler {
	return &Handler{log: log, service: service}
}

func AddHandler(ctx context.Context, log *slog.Logger, service LibraryService) func(r chi.Router) {
	handler := NewHandler(log, service)

	return func(r chi.Router) {
		r.Post("/save", handler.SaveSong(ctx))
		r.Post("/get", handler.GetLibrary(ctx))
		r.Get("/song-text", handler.GetSongText(ctx))
		r.Delete("/song/{id}", handler.DeleteSong(ctx))
		r.Patch("/update", handler.UpdateSong(ctx))
	}
}

// @Summary		Save a new song
// @Description	Save a new song into library.
// @Tags			API
// @Accept			json
// @Produce		json
// @Param			SongRequest	body		dto.SongRequest		true	"Song information"
// @Success		200			{object}	map[string]any		"success response"
// @Failure		500			{object}	map[string]string	"failure response"
// @Failure		400			{object}	map[string]string	"failure response"
// @Router			/save [post]
func (h *Handler) SaveSong(ctx context.Context) http.HandlerFunc {
	const op = "handlers.library.SaveSong"
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())

		h.log = with.WithOpAndRequestID(h.log, op, requestID)

		var song dto.SongRequest
		if err := render.Decode(r, &song); err != nil {
			h.log.Error("failed to decode model", sl.Err(err))
			handlers.ErrorResponse(w, r, 400, "failed to decode model")
			return
		}
		if err := song.Validate(); err != nil {
			h.log.Error("validation error in song info", sl.Err(err))
			handlers.ErrorResponse(w, r, 422, err.Error())
			return
		}

		id, err := h.service.SaveSong(ctx, song, requestID)
		if err != nil {
			h.log.Error("failed to save song", sl.Err(err))
			handlers.ErrorResponse(w, r, http.StatusInternalServerError, "failed to save song")
			return
		}

		handlers.SuccessResponse(w, r, http.StatusCreated, map[string]any{
			"detail": "new song successfully saved",
			"id":     id,
		})
	}
}

// @Summary		Get songs from library
// @Description	Get songs from library.
// @Tags			API
// @Accept			json
// @Produce		json
// @Param			Filters	body		dto.Filters			true	"Song information"
// @Param			limit	query		int					true	"limit"		default(10)
// @Param			offset	query		int					true	"offset"	default(0)
// @Success		200		{array}		models.Song			"success response"
// @Failure		500		{object}	map[string]string	"failure response"
// @Failure		400		{object}	map[string]string	"failure response"
// @Router			/get [post]
func (h *Handler) GetLibrary(ctx context.Context) http.HandlerFunc {
	const op = "handlers.library.GetLibrary"

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())

		h.log = with.WithOpAndRequestID(h.log, op, requestID)

		var filters dto.Filters
		if err := render.Decode(r, &filters); err != nil {
			h.log.Error("failed to decode filters", sl.Err(err))
			handlers.ErrorResponse(w, r, 400, "failed to decode filters")
			return
		}

		if err := filters.Validate(); err != nil {
			h.log.Error("validation error in filters", sl.Err(err))
			handlers.ErrorResponse(w, r, 422, err.Error())
			return
		}

		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil || limit <= 0 {
			limit = 10
		}

		offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
		if err != nil || offset < 0 {
			offset = 0
		}

		songs, err := h.service.GetLibrary(ctx, filters, limit, offset, requestID)
		if err != nil {
			h.log.Error("failed to get library", sl.Err(err))
			handlers.ErrorResponse(w, r, 400, err.Error())
			return
		}

		handlers.SuccessResponse(w, r, 200, songs)
	}
}

// @Summary		Get song text
// @Description	Get song text
// @Tags			API
// @Accept			json
// @Produce		json
// @Param			id		query		int					true	"songID"
// @Param			couplet	query		int					true	"couplet"	default(1)
// @Success		200		{object}	map[string]any		"success response"
// @Failure		500		{object}	map[string]string	"failure response"
// @Failure		400		{object}	map[string]string	"failure response"
// @Router			/song-text [get]
func (h *Handler) GetSongText(ctx context.Context) http.HandlerFunc {
	const op = "handlers.library.GetSongText"

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())

		h.log = with.WithOpAndRequestID(h.log, op, requestID)

		songID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil || songID <= 0 {
			h.log.Error("invalid song ID", sl.Err(err))
			handlers.ErrorResponse(w, r, 400, "invalid song ID")
			return
		}

		couplet, err := strconv.Atoi(r.URL.Query().Get("couplet"))
		if err != nil || couplet < 0 {
			couplet = 0
		}

		text, err := h.service.GetSongText(ctx, songID, couplet, requestID)
		if err != nil {
			h.log.Error("failed to get song text", sl.Err(err))
			handlers.ErrorResponse(w, r, 400, err.Error())
			return
		}

		handlers.SuccessResponse(w, r, 200, map[string]any{
			"song_id": songID,
			"couplet": couplet,
			"text":    text,
		})
	}
}

// @Summary		Delete song
// @Description	Delete song
// @Tags			API
// @Accept			json
// @Produce		json
// @Param			id	path		int					true	"songID"
// @Success		200	{object}	map[string]any		"success response"
// @Failure		500	{object}	map[string]string	"failure response"
// @Failure		400	{object}	map[string]string	"failure response"
// @Router			/song/{id} [delete]
func (h *Handler) DeleteSong(ctx context.Context) http.HandlerFunc {
	const op = "handlers.library.DeleteSong"

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())

		h.log = with.WithOpAndRequestID(h.log, op, requestID)

		songID, err := strconv.Atoi(chi.URLParam(r, "id"))
		if err != nil || songID <= 0 {
			h.log.Error("invalid song ID", sl.Err(err))
			handlers.ErrorResponse(w, r, 400, "invalid song ID")
			return
		}

		err = h.service.DeleteSong(ctx, songID, requestID)
		if err != nil {
			h.log.Error("failed to delete song", sl.Err(err))
			handlers.ErrorResponse(w, r, 400, err.Error())
			return
		}

		handlers.SuccessResponse(w, r, 200, map[string]any{
			"song_id": songID,
			"detail":  "song was successfully deleted",
		})
	}
}

// @Summary		Update song
// @Description	Update song
// @Tags			API
// @Accept			json
// @Produce		json
// @Param			UpdateSong	body		dto.UpdateSong		true	"Song information"
// @Success		200			{object}	map[string]any		"success response"
// @Failure		500			{object}	map[string]string	"failure response"
// @Failure		400			{object}	map[string]string	"failure response"
// @Router			/update [patch]
func (h *Handler) UpdateSong(ctx context.Context) http.HandlerFunc {
	const op = "handlers.library.UpdateSong"

	return func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())

		h.log = with.WithOpAndRequestID(h.log, op, requestID)

		var updateModel dto.UpdateSong
		if err := render.Decode(r, &updateModel); err != nil {
			h.log.Error("failed to decode update model", sl.Err(err))
			handlers.ErrorResponse(w, r, 400, "failed to decode update model")
			return
		}

		if err := updateModel.Validate(); err != nil {
			h.log.Error("validation error in update song info", sl.Err(err))
			handlers.ErrorResponse(w, r, 422, err.Error())
			return
		}

		if err := h.service.UpdateSong(ctx, updateModel, requestID); err != nil {
			h.log.Error("failed to update song", sl.Err(err))
			handlers.ErrorResponse(w, r, 400, err.Error())
			return
		}

		handlers.SuccessResponse(w, r, 200, map[string]any{
			"song_id": updateModel.ID,
			"detail":  "song successfully updated",
		})
	}
}
