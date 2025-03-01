package response

type SuccessResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
type ErrorCodeResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

type TeamResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InviteLink  string `json:"invitelink"`
}

type UserResponse struct {
	TelegramID string `json:"telegram_id"`
	Name       string `json:"name"`
}

type UserInfoResponse struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	TeamName string `json:"team_name"`
}
