package handlers

import (
	"github.com/labstack/echo/v4"
)

// Handler Struct for using methods for handling information.
type Handler struct {
}

func (h Handler) Bind(group *echo.Group) {
	group.GET("/read", h.GetReads)
	group.POST("/read/add", h.AddReads)
	group.DELETE("/read/delete", h.DeleteReads)
}

func (h Handler) BindRestricted(group *echo.Group) {
	// Key handlers
	group.POST("/key", h.GetKeys)
	group.POST("/key/add", h.AddKey)
	group.PUT("/key/update", h.UpdateKey)
	group.DELETE("/key/delete", h.DeleteKey)
}
