package api

import (
	"net/http"
	"ticketprocessing/internal/service"

	"github.com/labstack/echo/v4"
)

type RegisterRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type UserHandler struct {
	authService service.AuthService
}

func NewUserHandler(authService service.AuthService) *UserHandler {
	return &UserHandler{
		authService: authService,
	}
}

func (h *UserHandler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
	}

	err := h.authService.Register(c.Request().Context(), req.Name, req.Email, req.Password)
	switch err {
	case nil:
		return c.NoContent(http.StatusCreated)
	case service.ErrUserExists:
		return echo.NewHTTPError(http.StatusBadRequest, "user already exists")
	case service.ErrInternal:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "unexpected error")
	}
}

func (h *UserHandler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	token, err := h.authService.Login(c.Request().Context(), req.Email, req.Password)
	switch err {
	case nil:
		return c.JSON(http.StatusOK, AuthResponse{Token: token})
	case service.ErrInvalidCredentials:
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	case service.ErrInternal:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "unexpected error")
	}
}

func (h *UserHandler) Me(c echo.Context) error {
	userID := c.Get("user_id")
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Hello, user!",
		"user_id": userID,
	})
}
