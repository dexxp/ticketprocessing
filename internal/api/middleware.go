package api

import (
	"net/http"
	"strings"
	"ticketprocessing/internal/auth"

	"github.com/labstack/echo/v4"
)

func AuthMiddleware(jwtManager *auth.JWTManager, redisStore *auth.RedisTokenStore) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if header == "" || !strings.HasPrefix(header, "Bearer ") {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing token")
			}

			token := strings.TrimPrefix(header, "Bearer ")
			claims, err := jwtManager.Parse(token)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			if _, err := redisStore.GetUserID(c.Request().Context(), token); err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "token expired")
			}

			if err := redisStore.RefreshToken(c.Request().Context(), token); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to refresh token")
			}

			c.Set("user_id", claims.UserID)
			return next(c)
		}
	}
}
