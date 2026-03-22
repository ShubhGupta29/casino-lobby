package models

import "time"

// USER PROFILE
type UserProfile struct {
	UserID    string    `bson:"user_id" json:"user_id"`
	FirstName string    `bson:"first_name"`
	LastName  string    `bson:"last_name"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// USER AUTH
type UserAuth struct {
	UserID    string    `bson:"user_id"`
	Email     string    `bson:"email"`
	Hash      string    `bson:"hash"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// WALLET
type Wallet struct {
	UserID    string    `bson:"user_id"`
	Balance   string    `bson:"balance"`
	Currency  string    `bson:"currency"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// TRANSACTION
type Transaction struct {
	TransactionID string    `bson:"transaction_id"`
	UserID        string    `bson:"user_id"`
	Currency      string    `bson:"currency"`
	Amount        float64   `bson:"amount"`
	Type          string    `bson:"type"`
	RequestID     string    `bson:"request_id"`
	BalanceAfter  string    `bson:"balance_after"`
	CreatedAt     time.Time `bson:"created_at"`
	UpdatedAt     time.Time `bson:"updated_at"`
}

// SESSION
type Session struct {
	SessionID  string    `bson:"session_id"`
	UserID     string    `bson:"user_id"`
	OperatorID string    `bson:"operator_id"`
	GameID     string    `bson:"game_id"`
	Balance    string    `bson:"balance"`
	Currency   string    `bson:"currency"`
	Timestamp  int64     `bson:"timestamp"`
	CreatedAt  time.Time `bson:"created_at"`
	UpdatedAt  time.Time `bson:"updated_at"`
}

// ---------------- LOBBY API REQUESTS & RESPONSES ----------------

type RegisterRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// ---------------- OPERATOR API REQUESTS & RESPONSES ----------------

// AuthenticateRequest represents the request body for the Authenticate API.
type AuthenticateRequest struct {
	Token     string `json:"token"`
	RequestID string `json:"request_id"` // Note: API example uses "requestID", but docs table specifies "request_id".
}

// AuthenticateResponse represents the response body for the Authenticate API.
type AuthenticateResponse struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	Currency  string `json:"currency"`
	Balance   string `json:"balance"`
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
}

// BalanceRequest represents the request body for the Balance API.
type BalanceRequest struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	GameID    string `json:"game_id"`
	RequestID string `json:"request_id"`
}

// BalanceResponse represents the response body for the Balance API.
type BalanceResponse struct {
	UserID    string `json:"user_id"`
	Currency  string `json:"currency"`
	Balance   string `json:"balance"`
	RequestID string `json:"request_id"`
	Status    string `json:"status"`
}

// DebitRequest represents the request body for the Debit API.
type DebitRequest struct {
	SessionID      string  `json:"session_id"`
	UserID         string  `json:"user_id"`
	TransactionID  string  `json:"transaction_id"`
	Amount         float64 `json:"amount"`
	GameID         string  `json:"game_id"`
	Round          string  `json:"round"`
	RoundClosed    bool    `json:"round_closed"`
	Currency       string  `json:"currency"`
	RequestID      string  `json:"request_id"`
	IsFree         *bool   `json:"is_free,omitempty"`
	PfsCampaignID  string  `json:"pfs_campaign_id,omitempty"`
	PfsPromotionID string  `json:"pfs_promotion_id,omitempty"`
}

// DebitResponse represents the response body for the Debit API.
type DebitResponse struct {
	TransactionID string `json:"transaction_id"`
	Balance       string `json:"balance"`
	Currency      string `json:"currency"`
	RequestID     string `json:"request_id"`
	Status        string `json:"status"`
}

// CreditRequest represents the request body for the Credit API.
type CreditRequest struct {
	SessionID      string  `json:"session_id"`
	UserID         string  `json:"user_id"`
	TransactionID  string  `json:"transaction_id"`
	Amount         float64 `json:"amount"`
	GameID         string  `json:"game_id"`
	Round          string  `json:"round"`
	RoundClosed    bool    `json:"round_closed"`
	Currency       string  `json:"currency"`
	RequestID      string  `json:"request_id"`
	IsFree         *bool   `json:"is_free,omitempty"`
	PfsCampaignID  string  `json:"pfs_campaign_id,omitempty"`
	PfsPromotionID string  `json:"pfs_promotion_id,omitempty"`
	PfsCompleted   *bool   `json:"pfs_completed,omitempty"`
}

// CreditResponse represents the response body for the Credit API.
type CreditResponse struct {
	TransactionID string `json:"transaction_id"`
	Balance       string `json:"balance"`
	Currency      string `json:"currency"`
	RequestID     string `json:"request_id"`
	Status        string `json:"status"`
}

// RollbackRequest represents the request body for the Rollback API.
type RollbackRequest struct {
	TransactionID          string  `json:"transaction_id"`
	UserID                 string  `json:"user_id"`
	Amount                 float64 `json:"amount"`
	GameID                 string  `json:"game_id"`
	Reason                 string  `json:"reason,omitempty"`
	ReferenceTransactionID string  `json:"reference_transaction_id"`
	Round                  string  `json:"round"`
	RoundClosed            bool    `json:"round_closed"`
	SessionID              string  `json:"session_id"`
	RequestID              string  `json:"request_id"`
}

// RollbackResponse represents the response body for the Rollback API.
type RollbackResponse struct {
	TransactionID string `json:"transaction_id"`
	Balance       string `json:"balance"`
	RequestID     string `json:"request_id"`
	Status        string `json:"status"`
}
