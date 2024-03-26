package types

/*
	Responses
*/

type GetNotificationsResponse struct {
	ReaderName string       `json:"reader"`
	Note       Notification `json:"notification"`
}

/*
	Requests
*/

type SaveNotificationRequest struct {
	Note RequestNotification `json:"notification"`
}

type GetNotificationsRequest struct {
	ReaderName string `json:"reader"`
}
