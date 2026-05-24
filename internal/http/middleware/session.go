package middleware

import (
	"errors"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type SessionConfig struct {
	SessionManager *scs.SessionManager
	Skipper        middleware.Skipper
}

var DefaultSessionConfig = SessionConfig{
	Skipper: middleware.DefaultSkipper,
}

func Session(sessionManager *scs.SessionManager) echo.MiddlewareFunc {
	c := DefaultSessionConfig
	c.SessionManager = sessionManager
	return SessionWithConfig(c)
}

func SessionWithConfig(config SessionConfig) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultSessionConfig.Skipper
	}

	if config.SessionManager == nil {
		panic(errors.New("session manager is required"))
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			c.Response().Header().Set("Vary", "Cookie")

			s := config.SessionManager

			var token string
			cookie, err := c.Cookie(s.Cookie.Name)
			if err == nil {
				token = cookie.Value
			}

			ctx, err := s.Load(c.Request().Context(), token)
			if err != nil {
				return err
			}

			c.SetRequest(c.Request().WithContext(ctx))

			unwrapResp, err := echo.UnwrapResponse(c.Response())
			if err != nil {
				return err
			}

			unwrapResp.Before(func() {
				ctx := c.Request().Context()

				switch s.Status(ctx) {
				case scs.Modified:
					token, expiry, err := s.Commit(ctx)
					if err != nil {
						c.Logger().Error("failed to commit session", "error", err)
						return
					}

					s.WriteSessionCookie(ctx, c.Response(), token, expiry)
				case scs.Destroyed:
					s.WriteSessionCookie(ctx, c.Response(), "", time.Time{})
				default:
				}
			})

			return next(c)
		}
	}
}
