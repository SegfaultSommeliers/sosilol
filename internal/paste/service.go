package paste

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/SegfaultSommeliers/sosilol/internal/db"
	"github.com/SegfaultSommeliers/sosilol/internal/github"
	"github.com/SegfaultSommeliers/sosilol/internal/paste/cache"
	"github.com/SegfaultSommeliers/sosilol/internal/shared/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

var (
	ErrPasteNotFound = errors.New("paste not found")
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Service struct {
	queries *db.Queries
	logger  *slog.Logger

	githubService *github.Service
	cacheService  *cache.Service
}

func NewService(
	queries *db.Queries,
	logger *slog.Logger,

	githubService *github.Service,
	cacheService *cache.Service,
) *Service {
	return &Service{
		queries: queries,
		logger:  logger,

		githubService: githubService,
		cacheService:  cacheService,
	}
}

func (s *Service) Save(
	ctx context.Context,
	text string,
	token string,
) (string, error) {
	generatedID, err := generateID()
	if err != nil {
		return "", fmt.Errorf("failed to generate id: %w", err)
	}

	var authorID pgtype.Int8
	if token != "" {
		profile, err := s.githubService.GetRawProfile(ctx, token)
		if err != nil {
			return "", err
		}

		if err = s.queries.UpsertProfile(ctx, profile.ID); err != nil {
			return "", fmt.Errorf("failed to upsert profile: %w", err)
		}

		authorID = pgtype.Int8{
			Int64: profile.ID,
			Valid: true,
		}
	}

	err = s.queries.InsertPaste(ctx, db.InsertPasteParams{
		ID:       generatedID,
		Code:     text,
		AuthorID: authorID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to insert paste: %w", err)
	}

	return generatedID, nil
}

func (s *Service) GetPaste(ctx context.Context, id string) (*model.Paste, error) {
	code, err := s.cacheService.GetPaste(ctx, id)
	if err != nil {
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
		go func(bgCtx context.Context, pasteID, pasteCode string) {
			if cacheErr := s.cacheService.SetPaste(bgCtx, pasteID, pasteCode); cacheErr != nil {
				s.logger.ErrorContext(bgCtx, "failed to cache paste background",
					"id", pasteID,
					"error", cacheErr,
				)
			}
		}(context.WithoutCancel(ctx), id, code)
	}

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
