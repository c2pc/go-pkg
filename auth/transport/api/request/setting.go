package request

type SettingUpdateRequest struct {
	Settings *string `json:"settings" binding:"omitempty"`
}
