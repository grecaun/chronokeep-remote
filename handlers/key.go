package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h Handler) GetKeys(c echo.Context) error {
	return c.NoContent(http.StatusNotImplemented)
}

func (h Handler) AddKey(c echo.Context) error {
	return c.NoContent(http.StatusNotImplemented)
}

func (h Handler) UpdateKey(c echo.Context) error {
	return c.NoContent(http.StatusNotImplemented)
}

func (h Handler) DeleteKey(c echo.Context) error {
	return c.NoContent(http.StatusNotImplemented)
}
