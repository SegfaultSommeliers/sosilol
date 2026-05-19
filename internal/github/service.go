package github

import (
	"context"

	"github.com/SegfaultSommeliers/sosilol"
	"github.com/SegfaultSommeliers/sosilol/internal/db"
	"github.com/SegfaultSommeliers/sosilol/internal/shared/model"
	gogithub "github.com/google/go-github/v86/github"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

type Service struct {
	authConfig *oauth2.Config
	queries    *db.Queries
}

func NewService(
	clientId string,
	clientSecret string,
	queries *db.Queries,
) *Service {
	return &Service{
		authConfig: &oauth2.Config{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			Endpoint:     githuboauth.Endpoint,
		},
		queries: queries,
	}
}

func (s *Service) Authorize(
	ctx context.Context,
	code string,
) (string, error) {
	token, err := s.authConfig.Exchange(ctx, code)
	if err != nil {
		return "", sosilol.ErrExchangeCodeFailed
	}

	return token.AccessToken, nil
}

func (s *Service) GetRawProfile(
	ctx context.Context,
	token string,
) (*model.Profile, error) {
	client, err := gogithub.NewClient(gogithub.WithAuthToken(token))
	if err != nil {
		return nil, sosilol.ErrGetGithubClientFailed
	}

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, sosilol.ErrUserNotFound
	}

	return &model.Profile{
		ID:        user.GetID(),
		Login:     user.GetLogin(),
		AvatarURL: user.GetAvatarURL(),
		Pastes:    make([]model.Paste, 0),
	}, nil
}

func (s *Service) GetProfile(
	ctx context.Context,
	token string,
) (*model.Profile, error) {
	client, err := gogithub.NewClient(gogithub.WithAuthToken(token))
	if err != nil {
		return nil, sosilol.ErrGetGithubClientFailed
	}

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, sosilol.ErrUserNotFound
	}

	dbPastes, err := s.queries.GetPastesByAuthorID(ctx, pgtype.Int8{
		Int64: user.GetID(),
		Valid: true,
	})
	if err != nil {
		return nil, err
	}

	pastes := make([]model.Paste, len(dbPastes))
	for i, dbPaste := range dbPastes {
		pastes[i] = model.Paste{
			ID:   dbPaste.ID,
			Code: dbPaste.Code,
		}
	}

	return &model.Profile{
		ID:        user.GetID(),
		Login:     user.GetLogin(),
		AvatarURL: user.GetAvatarURL(),
		Pastes:    pastes,
	}, nil
}
