package paste

import (
	"context"
	"errors"

	"github.com/SegfaultSommeliers/sosilol"
	"github.com/SegfaultSommeliers/sosilol/internal/db"
	"github.com/SegfaultSommeliers/sosilol/internal/github"
	"github.com/SegfaultSommeliers/sosilol/internal/paste/cache"
	"github.com/SegfaultSommeliers/sosilol/internal/shared/model"
	"github.com/SegfaultSommeliers/sosilol/internal/shared/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	queries *db.Queries

	githubService *github.Service
	cacheService  *cache.Service
}

func NewService(
	queries *db.Queries,

	githubService *github.Service,
	cacheService *cache.Service,
) *Service {
	return &Service{
		queries: queries,

		githubService: githubService,
		cacheService:  cacheService,
	}
}

func (s *Service) Save(
	ctx context.Context,
	text string,
	token string,
) (string, error) {
	generatedID, err := utils.RandomID(7)
	if err != nil {
		return "", err
	}

	if token != "" {
		profile, err := s.githubService.GetProfile(ctx, token)
		if err != nil {
			return "", err
		}

		if err = s.queries.UpsertProfile(ctx, profile.ID); err != nil {
			return "", err
		}

		err = s.queries.InsertPaste(ctx, db.InsertPasteParams{
			ID:   generatedID,
			Code: text,
			AuthorID: pgtype.Int8{
				Int64: profile.ID,
				Valid: true,
			},
		})

		if err != nil {
			return "", err
		}
		return generatedID, nil
	}

	err = s.queries.InsertPaste(ctx, db.InsertPasteParams{
		ID:   generatedID,
		Code: text,
	})
	if err != nil {
		return "", err
	}

	return generatedID, nil
}

func (s *Service) GetPaste(ctx context.Context, id string) (*model.Paste, error) {
	code, err := s.cacheService.GetPaste(ctx, id)
	if err != nil {
		if !errors.Is(err, cache.ErrCacheMiss) {
			return nil, err
		}

		paste, err := s.queries.GetPaste(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, sosilol.ErrPasteNotFound
			}

			return nil, err
		}
		code = paste.Code
		_ = s.cacheService.SetPaste(ctx, id, code)
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
