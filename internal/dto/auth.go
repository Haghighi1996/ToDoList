package dto

type AnonymousAuthRequest struct {
	DeviceID string `json:"device_id"`
}

type AnonymousAuthResponse struct {
	Token     string `json:"token"`
	UserID    string `json:"user_id"`
	ExpiresAt string `json:"expires_at"` // ISO 8601 string
}
