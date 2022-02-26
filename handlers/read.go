package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h Handler) GetReads(c echo.Context) error {
	return c.NoContent(http.StatusNotImplemented)
}

func (h Handler) AddReads(c echo.Context) error {
	return c.NoContent(http.StatusNotImplemented)
}

func (h Handler) DeleteReads(c echo.Context) error {
	return c.NoContent(http.StatusNotImplemented)
}
