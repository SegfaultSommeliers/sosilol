package paste

import (
	"context"
	"errors"

	"github.com/SegfaultSommeliers/sosilol"
	"github.com/SegfaultSommeliers/sosilol/internal/db"
	"github.com/SegfaultSommeliers/sosilol/internal/github"
	"github.com/SegfaultSommeliers/sosilol/internal/shared/model"
	"github.com/SegfaultSommeliers/sosilol/internal/shared/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Service struct {
	queries *db.Queries

	githubService *github.Service
}

func NewService(
	queries *db.Queries,

	githubService *github.Service,
) *Service {
	return &Service{
		queries: queries,

		githubService: githubService,
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
	paste, err := s.queries.GetPaste(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sosilol.ErrPasteNotFound
		}

		return nil, err
	}

	return &model.Paste{
		ID:   paste.ID,
		Code: paste.Code,
	}, nil
}

func (s *Service) LoadRaw(ctx context.Context, id string) (string, error) {
	paste, err := s.GetPaste(ctx, id)
	if err != nil {
		return "", err
	}

	return paste.Code, nil
}
