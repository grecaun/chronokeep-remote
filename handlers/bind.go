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
	// Account handlers
	group.POST("/account", h.GetAccount)
	group.GET("/account/all", h.GetAccounts)
	group.POST("/account/logout", h.Logout)
	group.POST("/account/add", h.AddAccount)
	group.PUT("/account/update", h.UpdateAccount)
	group.PUT("/account/password", h.ChangePassword)
	group.PUT("/account/email", h.ChangeEmail)
	group.POST("/account/unlock", h.Unlock)
	group.DELETE("/account/delete", h.DeleteAccount)
	// Key handlers
	group.POST("/key", h.GetKeys)
	group.POST("/key/add", h.AddKey)
	group.PUT("/key/update", h.UpdateKey)
	group.DELETE("/key/delete", h.DeleteKey)
}
