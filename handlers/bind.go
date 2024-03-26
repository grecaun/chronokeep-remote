package handlers

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// Handler Struct for using methods for handling information.
type Handler struct {
	validate *validator.Validate
}

func (h Handler) Bind(group *echo.Group) {
	// Read handlers
	group.GET("/reads", h.GetReads)
	group.POST("/reads/add", h.AddReads)
	group.DELETE("/reads/delete", h.DeleteReads)
	// Reader handler(s)
	group.GET("/readers", h.GetReaders)
	// Account Login
	group.POST("/account/login", h.Login)
	group.POST("/account/refresh", h.Refresh)
	// Notification handlers
	group.POST("/notifications/save", h.SaveNotification)
	group.GET("/notifications/get", h.GetNotifications)
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
