package paste

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/SegfaultSommeliers/sosilol/internal/db"
	"github.com/SegfaultSommeliers/sosilol/internal/paste/cache"
	"github.com/SegfaultSommeliers/sosilol/internal/shared/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

var ErrPasteNotFound = errors.New("paste not found")

const (
	alphabet   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	maxRetries = 5
)

type Service struct {
	queries *db.Queries
	logger  *slog.Logger

	cacheService *cache.Service
}

func NewService(
	queries *db.Queries,
	logger *slog.Logger,

	cacheService *cache.Service,
) *Service {
	return &Service{
		queries: queries,
		logger:  logger,

		cacheService: cacheService,
	}
}

func (s *Service) Save(
	ctx context.Context,
	text string,
	authorUserID int64,
) (string, error) {
	var authorID pgtype.Int8
	if authorUserID != 0 {
		if err := s.queries.UpsertProfile(ctx, authorUserID); err != nil {
			return "", fmt.Errorf("failed to upsert profile: %w", err)
		}
		authorID = pgtype.Int8{
			Int64: authorUserID,
			Valid: true,
		}
	}

	for range maxRetries {
		generatedID, err := generateID()
		if err != nil {
			return "", fmt.Errorf("failed to generate id: %w", err)
		}

		rowsAffected, err := s.queries.InsertPaste(ctx, db.InsertPasteParams{
			ID:       generatedID,
			Code:     text,
			AuthorID: authorID,
		})
		if err != nil {
			return "", fmt.Errorf("failed to insert paste: %w", err)
		}

		if rowsAffected > 0 {
			return generatedID, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique id after %d attempts", maxRetries)
}

func (s *Service) GetPaste(ctx context.Context, id string) (*model.Paste, error) {
	code, err := s.cacheService.GetPaste(ctx, id)
	if err == nil {
		return &model.Paste{
			ID:   id,
			Code: code,
		}, nil
	}

	if !errors.Is(err, cache.ErrCacheMiss) {
		s.logger.ErrorContext(
			ctx,
			"failed to get paste from cache",
			"id", id,
			"error", err,
		)
	}

	paste, err := s.queries.GetPaste(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPasteNotFound
		}

		return nil, err
	}
	code = paste.Code
	go func(id, code string) {
		cacheCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 200*time.Millisecond)
		defer cancel()

		if cacheErr := s.cacheService.SetPaste(cacheCtx, id, code); cacheErr != nil {
			s.logger.ErrorContext(cacheCtx, "failed to cache paste background",
				"id", id,
				"error", cacheErr,
			)
		}
	}(id, code)

	return &model.Paste{
		ID:   id,
		Code: code,
	}, nil
}

func (s *Service) LoadRaw(ctx context.Context, id string) (string, error) {
	paste, err := s.GetPaste(ctx, id)
	if err != nil {
		return "", err
	}

	return paste.Code, nil
}

func generateID() (string, error) {
	return gonanoid.Generate(alphabet, 7)
}
