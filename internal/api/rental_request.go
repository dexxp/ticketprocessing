package api

import (
	"net/http"
	"strconv"
	"ticketprocessing/internal/service"
	"time"

	"github.com/labstack/echo/v4"
)

type RentalRequestHandler struct {
	rentalRequestService service.RentalRequestService
}

func NewRentalRequestHandler(rentalRequestService service.RentalRequestService) *RentalRequestHandler {
	return &RentalRequestHandler{
		rentalRequestService: rentalRequestService,
	}
}

// CreateRentalRequest godoc
// @Summary Create a new rental request
// @Description Create a new rental request and send it to the queue for processing
// @Tags rental-requests
// @Accept json
// @Produce json
// @Param request body service.CreateRentalRequestRequest true "Rental Request"
// @Success 201 {object} models.RentalRequest
// @Failure 400 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Security BearerAuth
// @Router /rental_request [post]
func (h *RentalRequestHandler) CreateRentalRequest(c echo.Context) error {
	var req service.CreateRentalRequestRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Получаем ID пользователя из контекста (установлен middleware аутентификации)
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid user context")
	}

	request, err := h.rentalRequestService.CreateRentalRequest(c.Request().Context(), userID, req)
	switch err {
	case nil:
		return c.JSON(http.StatusCreated, request)
	case service.ErrEquipmentNotFound:
		return echo.NewHTTPError(http.StatusNotFound, "equipment not found")
	case service.ErrInvalidDateRange:
		return echo.NewHTTPError(http.StatusBadRequest, "invalid date range")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}

// GetRequestStatus godoc
// @Summary Get current status of a rental request
// @Description Get the latest status of a rental request by its ID
// @Tags rental-requests
// @Accept json
// @Produce json
// @Param id path int true "Rental Request ID"
// @Success 200 {object} models.RequestStatusLog
// @Failure 400 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Security BearerAuth
// @Router /rental_request/{id}/status [get]
func (h *RentalRequestHandler) GetRequestStatus(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request id")
	}

	status, err := h.rentalRequestService.GetRequestStatus(c.Request().Context(), uint(id))
	switch err {
	case nil:
		return c.JSON(http.StatusOK, status)
	case service.ErrRentalRequestNotFound:
		return echo.NewHTTPError(http.StatusNotFound, "rental request not found")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}

// GetRequestStatusAt godoc
// @Summary Get status of a rental request at specific time
// @Description Get the status of a rental request at a specific point in time
// @Tags rental-requests
// @Accept json
// @Produce json
// @Param id path int true "Rental Request ID"
// @Param datetime query string true "DateTime in RFC3339 format" format(date-time)
// @Success 200 {object} models.RequestStatusLog
// @Failure 400 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Security BearerAuth
// @Router /rental_request/{id}/status_at [get]
func (h *RentalRequestHandler) GetRequestStatusAt(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request id")
	}

	datetimeStr := c.QueryParam("datetime")
	if datetimeStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "datetime parameter is required")
	}

	datetime, err := time.Parse(time.RFC3339, datetimeStr)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid datetime format, use RFC3339")
	}

	status, err := h.rentalRequestService.GetRequestStatusAt(c.Request().Context(), uint(id), datetime)
	switch err {
	case nil:
		return c.JSON(http.StatusOK, status)
	case service.ErrRentalRequestNotFound:
		return echo.NewHTTPError(http.StatusNotFound, "rental request not found")
	case service.ErrInvalidDateTime:
		return echo.NewHTTPError(http.StatusBadRequest, "invalid datetime")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}
}
