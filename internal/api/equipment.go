package api

import (
	"net/http"
	"strconv"
	"ticketprocessing/internal/models"
	"ticketprocessing/internal/service"

	"github.com/labstack/echo/v4"
)

type EquipmentHandler struct {
	service *service.EquipmentService
}

func NewEquipmentHandler(service *service.EquipmentService) *EquipmentHandler {
	return &EquipmentHandler{
		service: service,
	}
}

func (h *EquipmentHandler) RegisterRoutes(e *echo.Echo) {
	equipment := e.Group("/api/equipment")
	equipment.POST("", h.Create)
	equipment.GET("/:id", h.GetByID)
	equipment.PUT("/:id", h.Update)
	equipment.DELETE("/:id", h.Delete)
}

func (h *EquipmentHandler) Create(c echo.Context) error {
	var equipment models.Equipment
	if err := c.Bind(&equipment); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := h.service.Create(c.Request().Context(), &equipment); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, equipment)
}

func (h *EquipmentHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	equipment, err := h.service.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "equipment not found"})
	}

	return c.JSON(http.StatusOK, equipment)
}

func (h *EquipmentHandler) Update(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	var equipment models.Equipment
	if err := c.Bind(&equipment); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	equipment.ID = uint(id)
	if err := h.service.Update(c.Request().Context(), &equipment); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, equipment)
}

func (h *EquipmentHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	equipment := &models.Equipment{ID: uint(id)}
	if err := h.service.Delete(c.Request().Context(), equipment); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}
