package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/SegfaultSommeliers/sosilol/internal/db"
	"github.com/SegfaultSommeliers/sosilol/internal/shared/model"
	gogithub "github.com/google/go-github/v86/github"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"
)

func safeAvatarURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" {
		return ""
	}
	host := strings.ToLower(u.Host)
	if host != "avatars.githubusercontent.com" && host != "github.com" {
		return ""
	}
	return raw
}

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrUserNotFound = errors.New("user not found")
)

type Service struct {
	authConfig *oauth2.Config
	queries    *db.Queries
}

func NewService(
	clientId string,
	clientSecret string,
	redirectURL string,
	queries *db.Queries,
) *Service {
	return &Service{
		authConfig: &oauth2.Config{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Endpoint:     githuboauth.Endpoint,
			Scopes: []string{
				"read:user",
			},
		},
		queries: queries,
	}
}

func (s *Service) GetAuthURL(state string) string {
	return s.authConfig.AuthCodeURL(state)
}

func (s *Service) Authorize(
	ctx context.Context,
	code string,
) (string, error) {
	token, err := s.authConfig.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("failed to exchange code for token: %w", err)
	}

	return token.AccessToken, nil
}

func (s *Service) GetRawProfile(
	ctx context.Context,
	token string,
) (*model.Profile, error) {
	client, err := gogithub.NewClient(gogithub.WithAuthToken(token))
	if err != nil {
		if errResponse, ok := errors.AsType[*gogithub.ErrorResponse](err); ok {
			switch errResponse.Response.StatusCode {
			case http.StatusUnauthorized:
				return nil, ErrUnauthorized
			}
		}

		return nil, fmt.Errorf("failed to create GitHub client: %w", err)
	}

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		if errResponse, ok := errors.AsType[*gogithub.ErrorResponse](err); ok {
			switch errResponse.Response.StatusCode {
			case http.StatusUnauthorized:
				return nil, ErrUnauthorized
			case http.StatusNotFound:
				return nil, ErrUserNotFound
			}
		}

		return nil, fmt.Errorf("failed to get user info from GitHub: %w", err)
	}

	return &model.Profile{
		ID:        user.GetID(),
		Login:     user.GetLogin(),
		AvatarURL: safeAvatarURL(user.GetAvatarURL()),
		Pastes:    make([]model.Paste, 0),
	}, nil
}

func (s *Service) GetPastesByUserID(
	ctx context.Context,
	userID int64,
) ([]model.Paste, error) {
	dbPastes, err := s.queries.GetPastesByAuthorID(ctx, pgtype.Int8{
		Int64: userID,
		Valid: true,
	})
	if err != nil {
		return nil, err
	}

	pastes := make([]model.Paste, len(dbPastes))
	for i, dbPaste := range dbPastes {
		pastes[i] = model.Paste{
			ID: dbPaste.ID,
		}
	}
	return pastes, nil
}
