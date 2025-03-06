package response

import "time"

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

type MeetingResponse struct {
	ID             uint      `json:"id"`
	Title          string    `json:"title"`
	MeetingType    string    `json:"meeting_type"`
	Date           time.Time `json:"date"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	ConferenceLink string    `json:"conference_link"`
	Room           string    `json:"room"`
	TeamID         uint      `json:"team_id"`
	CreatedBy      uint      `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type TaskResponse struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Deadline    time.Time `json:"deadline"`
	Status      string    `json:"status"`
	IsTeam      bool      `json:"is_team"`
	AssignedTo  *string   `json:"assigned_to"`
	CreatedBy   uint      `json:"created_by"`
	TeamID      uint      `json:"team_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
