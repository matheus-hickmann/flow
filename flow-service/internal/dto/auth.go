// Package dto holds request/response shapes shared by handlers and services.
package dto

// SignupRequest is the body of POST /api/v1/auth/signup.
type SignupRequest struct {
	UserID      string `json:"userId"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName,omitempty"`
}

// LoginRequest is the body of POST /api/v1/auth/login.
type LoginRequest struct {
	UserID   string `json:"userId"`
	Password string `json:"password"`
}

// AuthResponse is returned on successful signup/login.
type AuthResponse struct {
	AccessToken string `json:"accessToken"`
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName,omitempty"`
}

// MeResponse is the body of GET /api/v1/users/me.
type MeResponse struct {
	UserID      string `json:"userId"`
	Source      string `json:"source"`
	Name        string `json:"name,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
}

// RecoveryQuestionItem is one question + answer (answer stored, not returned).
type RecoveryQuestionItem struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// CreateRecoveryQuestionsRequest must contain exactly 3 items.
type CreateRecoveryQuestionsRequest struct {
	Questions []RecoveryQuestionItem `json:"questions"`
}
