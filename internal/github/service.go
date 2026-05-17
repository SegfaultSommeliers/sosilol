package github

import (
	"context"

	"github.com/SegfaultSommeliers/sosilol/internal/db"
	"github.com/SegfaultSommeliers/sosilol/internal/paste"
	gogithub "github.com/google/go-github/v86/github"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

type Service struct {
	authConfig *oauth2.Config
	dbPool     *pgxpool.Pool
}

func NewService(
	clientId string,
	clientSecret string,
	dbPool *pgxpool.Pool,
) *Service {
	return &Service{
		authConfig: &oauth2.Config{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			Endpoint:     githuboauth.Endpoint,
		},
		dbPool: dbPool,
	}
}

func (s *Service) Authorize(
	ctx context.Context,
	code string,
) (string, error) {
	token, err := s.authConfig.Exchange(ctx, code)
	if err != nil {
		return "", err
	}

	return token.AccessToken, nil
}

func (s *Service) GetRawProfile(
	ctx context.Context,
	token string,
) (map[string]any, error) {
	client, err := gogithub.NewClient(gogithub.WithAuthToken(token))
	if err != nil {
		return nil, err
	}

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"id":         user.GetID(),
		"login":      user.GetLogin(),
		"avatar_url": user.GetAvatarURL(),
		"pastes":     make([]paste.Paste, 0),
	}, nil
}

func (s *Service) GetProfile(
	ctx context.Context,
	token string,
) (map[string]any, error) {
	client, err := gogithub.NewClient(gogithub.WithAuthToken(token))
	if err != nil {
		return nil, err
	}

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	q := db.New(s.dbPool)
	dbPastes, err := q.GetPastesByAuthorID(ctx, pgtype.Int8{
		Int64: user.GetID(),
		Valid: true,
	})

	pastes := make([]paste.Paste, len(dbPastes))
	for i, dbPaste := range dbPastes {
		pastes[i] = paste.Paste{
			ID:   dbPaste.ID,
			Code: dbPaste.Code,
		}
	}

	return map[string]any{
		"id":         user.GetID(),
		"login":      user.GetLogin(),
		"avatar_url": user.GetAvatarURL(),
		"pastes":     pastes,
	}, nil
}
