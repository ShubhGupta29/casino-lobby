package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"CasinoLobbyBE/db"
	models "CasinoLobbyBE/types"
	"CasinoLobbyBE/utils/response"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Authenticate handles the POST /generic/v1/operator/authenticate/{partner_id} endpoint.
func Authenticate(w http.ResponseWriter, r *http.Request) {
	partnerID := r.PathValue("partner_id")
	if partnerID == "" {
		response.GeneralErrorResponse(w, http.StatusBadRequest, "partner_id is required")
		return
	}

	var req models.AuthenticateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.GeneralErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var missing []string
	if req.Token == "" {
		missing = append(missing, "token")
	}
	if req.RequestID == "" {
		missing = append(missing, "request_id")
	}
	if len(missing) > 0 {
		response.GeneralErrorResponse(w, http.StatusBadRequest, "Missing required fields: "+strings.Join(missing, ", "))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Validate the JWT token to extract the real user_id
	userID, err := ValidateToken(req.Token)
	if err != nil {
		response.GeneralErrorResponse(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}
	sessionID := "sess_" + req.RequestID

	walletCol := db.GetCollection("wallet")
	var wallet models.Wallet
	err = walletCol.FindOne(ctx, bson.M{"user_id": userID}).Decode(&wallet)
	if err == mongo.ErrNoDocuments {
		wallet = models.Wallet{
			UserID:    userID,
			Balance:   "10000.00", // Assign a default mock balance
			Currency:  "EUR",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_, _ = walletCol.InsertOne(ctx, wallet)
	} else if err != nil {
		response.GeneralErrorResponse(w, http.StatusInternalServerError, "Database error")
		return
	}

	session := models.Session{
		SessionID:  sessionID,
		UserID:     userID,
		OperatorID: partnerID,
		Balance:    wallet.Balance,
		Currency:   wallet.Currency,
		Timestamp:  time.Now().Unix(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	_, _ = db.GetCollection("session").InsertOne(ctx, session)

	res := models.AuthenticateResponse{
		UserID:    userID,
		SessionID: sessionID,
		Currency:  wallet.Currency,
		Balance:   wallet.Balance,
		RequestID: req.RequestID,
		Status:    "OK",
	}

	response.WriteJSONResponse(w, http.StatusOK, res)
}

// Balance handles the POST /generic/v1/operator/balance/{partner_id} endpoint.
func Balance(w http.ResponseWriter, r *http.Request) {
	partnerID := r.PathValue("partner_id")
	if partnerID == "" {
		response.GeneralErrorResponse(w, http.StatusBadRequest, "partner_id is required")
		return
	}

	var req models.BalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.GeneralErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var missing []string
	if req.UserID == "" {
		missing = append(missing, "user_id")
	}
	if req.SessionID == "" {
		missing = append(missing, "session_id")
	}
	if req.GameID == "" {
		missing = append(missing, "game_id")
	}
	if req.RequestID == "" {
		missing = append(missing, "request_id")
	}
	if len(missing) > 0 {
		response.GeneralErrorResponse(w, http.StatusBadRequest, "Missing required fields: "+strings.Join(missing, ", "))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wallet models.Wallet
	err := db.GetCollection("wallet").FindOne(ctx, bson.M{"user_id": req.UserID}).Decode(&wallet)
	if err != nil {
		response.GeneralErrorResponse(w, http.StatusNotFound, "Wallet not found")
		return
	}

	res := models.BalanceResponse{
		UserID:    req.UserID,
		Currency:  wallet.Currency,
		Balance:   wallet.Balance,
		RequestID: req.RequestID,
		Status:    "OK",
	}

	response.WriteJSONResponse(w, http.StatusOK, res)
}

// Debit handles the POST /generic/v1/operator/debit/{partner_id} endpoint.
func Debit(w http.ResponseWriter, r *http.Request) {
	partnerID := r.PathValue("partner_id")
	if partnerID == "" {
		response.GeneralErrorResponse(w, http.StatusBadRequest, "partner_id is required")
		return
	}

	var req models.DebitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.GeneralErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var missing []string
	if req.UserID == "" {
		missing = append(missing, "user_id")
	}
	if req.SessionID == "" {
		missing = append(missing, "session_id")
	}
	if req.TransactionID == "" {
		missing = append(missing, "transaction_id")
	}
	if req.GameID == "" {
		missing = append(missing, "game_id")
	}
	if req.Round == "" {
		missing = append(missing, "round")
	}
	if req.Currency == "" {
		missing = append(missing, "currency")
	}
	if req.RequestID == "" {
		missing = append(missing, "request_id")
	}
	if len(missing) > 0 {
		response.GeneralErrorResponse(w, http.StatusBadRequest, "Missing required fields: "+strings.Join(missing, ", "))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Idempotency Check: if this debit was already recorded, return previous success
	var existingTx models.Transaction
	err := db.GetCollection("transaction").FindOne(ctx, bson.M{"transaction_id": req.TransactionID}).Decode(&existingTx)
	if err == nil {
		res := models.DebitResponse{
			Currency:      existingTx.Currency,
			Balance:       existingTx.BalanceAfter,
			Status:        "OK",
			RequestID:     existingTx.RequestID,
			TransactionID: existingTx.TransactionID,
		}
		response.WriteJSONResponse(w, http.StatusOK, res)
		return
	}

	var wallet models.Wallet
	err = db.GetCollection("wallet").FindOne(ctx, bson.M{"user_id": req.UserID}).Decode(&wallet)
	if err == mongo.ErrNoDocuments {
		res := models.DebitResponse{Status: "INSUFFICIENT_FUNDS", RequestID: req.RequestID}
		response.WriteJSONResponse(w, http.StatusOK, res)
		return
	} else if err != nil {
		response.GeneralErrorResponse(w, http.StatusInternalServerError, "Database error")
		return
	}

	balanceFloat, err := strconv.ParseFloat(wallet.Balance, 64)
	if err != nil {
		response.GeneralErrorResponse(w, http.StatusInternalServerError, "Invalid balance format")
		return
	}

	if balanceFloat < req.Amount {
		res := models.DebitResponse{Status: "INSUFFICIENT_FUNDS", RequestID: req.RequestID}
		response.WriteJSONResponse(w, http.StatusOK, res)
		return
	}

	newBalance := balanceFloat - req.Amount
	newBalanceStr := fmt.Sprintf("%.2f", newBalance)

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedWallet models.Wallet
	err = db.GetCollection("wallet").FindOneAndUpdate(
		ctx,
		bson.M{"user_id": req.UserID},
		bson.M{"$set": bson.M{"balance": newBalanceStr, "updated_at": time.Now()}},
		opts,
	).Decode(&updatedWallet)

	if err != nil {
		response.GeneralErrorResponse(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Record transaction
	db.GetCollection("transaction").InsertOne(ctx, models.Transaction{
		TransactionID: req.TransactionID,
		UserID:        req.UserID,
		Currency:      req.Currency,
		Amount:        req.Amount,
		Type:          "debit",
		RequestID:     req.RequestID,
		BalanceAfter:  updatedWallet.Balance,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})

	res := models.DebitResponse{
		Currency:      req.Currency,
		Balance:       updatedWallet.Balance,
		Status:        "OK",
		RequestID:     req.RequestID,
		TransactionID: req.TransactionID,
	}

	response.WriteJSONResponse(w, http.StatusOK, res)
}

// Credit handles the POST /generic/v1/operator/credit/{partner_id} endpoint.
func Credit(w http.ResponseWriter, r *http.Request) {
	partnerID := r.PathValue("partner_id")
	if partnerID == "" {
		response.GeneralErrorResponse(w, http.StatusBadRequest, "partner_id is required")
		return
	}

	var req models.CreditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.GeneralErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var missing []string
	if req.UserID == "" {
		missing = append(missing, "user_id")
	}
	if req.SessionID == "" {
		missing = append(missing, "session_id")
	}
	if req.TransactionID == "" {
		missing = append(missing, "transaction_id")
	}
	if req.GameID == "" {
		missing = append(missing, "game_id")
	}
	if req.Round == "" {
		missing = append(missing, "round")
	}
	if req.Currency == "" {
		missing = append(missing, "currency")
	}
	if req.RequestID == "" {
		missing = append(missing, "request_id")
	}
	if len(missing) > 0 {
		response.GeneralErrorResponse(w, http.StatusBadRequest, "Missing required fields: "+strings.Join(missing, ", "))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Idempotency Check
	var existingTx models.Transaction
	err := db.GetCollection("transaction").FindOne(ctx, bson.M{"transaction_id": req.TransactionID}).Decode(&existingTx)
	if err == nil {
		res := models.CreditResponse{
			Currency:      existingTx.Currency,
			Balance:       existingTx.BalanceAfter,
			Status:        "OK",
			RequestID:     existingTx.RequestID,
			TransactionID: existingTx.TransactionID,
		}
		response.WriteJSONResponse(w, http.StatusOK, res)
		return
	}

	var wallet models.Wallet
	err = db.GetCollection("wallet").FindOne(ctx, bson.M{"user_id": req.UserID}).Decode(&wallet)
	if err != nil {
		response.GeneralErrorResponse(w, http.StatusInternalServerError, "Database error")
		return
	}

	balanceFloat, err := strconv.ParseFloat(wallet.Balance, 64)
	if err != nil {
		response.GeneralErrorResponse(w, http.StatusInternalServerError, "Invalid balance format")
		return
	}

	newBalance := balanceFloat + req.Amount
	newBalanceStr := fmt.Sprintf("%.2f", newBalance)

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedWallet models.Wallet
	err = db.GetCollection("wallet").FindOneAndUpdate(
		ctx,
		bson.M{"user_id": req.UserID},
		bson.M{"$set": bson.M{"balance": newBalanceStr, "updated_at": time.Now()}},
		opts,
	).Decode(&updatedWallet)

	if err != nil {
		response.GeneralErrorResponse(w, http.StatusInternalServerError, "Database error")
		return
	}

	// Record transaction
	db.GetCollection("transaction").InsertOne(ctx, models.Transaction{
		TransactionID: req.TransactionID,
		UserID:        req.UserID,
		Currency:      req.Currency,
		Amount:        req.Amount,
		Type:          "credit",
		RequestID:     req.RequestID,
		BalanceAfter:  updatedWallet.Balance,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})

	res := models.CreditResponse{
		Currency:      req.Currency,
		Balance:       updatedWallet.Balance,
		Status:        "OK",
		RequestID:     req.RequestID,
		TransactionID: req.TransactionID,
	}

	response.WriteJSONResponse(w, http.StatusOK, res)
}

// Rollback handles the POST /generic/v1/operator/rollback/{partner_id} endpoint.
func Rollback(w http.ResponseWriter, r *http.Request) {
	partnerID := r.PathValue("partner_id")
	if partnerID == "" {
		response.GeneralErrorResponse(w, http.StatusBadRequest, "partner_id is required")
		return
	}

	var req models.RollbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.GeneralErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var missing []string
	if req.TransactionID == "" {
		missing = append(missing, "transaction_id")
	}
	if req.UserID == "" {
		missing = append(missing, "user_id")
	}
	if req.GameID == "" {
		missing = append(missing, "game_id")
	}
	if req.ReferenceTransactionID == "" {
		missing = append(missing, "reference_transaction_id")
	}
	if req.Round == "" {
		missing = append(missing, "round")
	}
	if req.SessionID == "" {
		missing = append(missing, "session_id")
	}
	if req.RequestID == "" {
		missing = append(missing, "request_id")
	}
	if len(missing) > 0 {
		response.GeneralErrorResponse(w, http.StatusBadRequest, "Missing required fields: "+strings.Join(missing, ", "))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Idempotency Check
	var existingTx models.Transaction
	err := db.GetCollection("transaction").FindOne(ctx, bson.M{"transaction_id": req.TransactionID}).Decode(&existingTx)
	if err == nil {
		res := models.RollbackResponse{
			Balance:       existingTx.BalanceAfter,
			Status:        "OK",
			RequestID:     existingTx.RequestID,
			TransactionID: existingTx.TransactionID,
		}
		response.WriteJSONResponse(w, http.StatusOK, res)
		return
	}

	var wallet models.Wallet
	err = db.GetCollection("wallet").FindOne(ctx, bson.M{"user_id": req.UserID}).Decode(&wallet)
	if err != nil {
		response.GeneralErrorResponse(w, http.StatusInternalServerError, "Database error")
		return
	}

	balanceFloat, err := strconv.ParseFloat(wallet.Balance, 64)
	if err != nil {
		response.GeneralErrorResponse(w, http.StatusInternalServerError, "Invalid balance format")
		return
	}

	newBalance := balanceFloat + req.Amount
	newBalanceStr := fmt.Sprintf("%.2f", newBalance)

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedWallet models.Wallet
	err = db.GetCollection("wallet").FindOneAndUpdate(
		ctx,
		bson.M{"user_id": req.UserID},
		bson.M{"$set": bson.M{"balance": newBalanceStr, "updated_at": time.Now()}},
		opts,
	).Decode(&updatedWallet)

	if err != nil {
		response.GeneralErrorResponse(w, http.StatusInternalServerError, "Database error")
		return
	}

	db.GetCollection("transaction").InsertOne(ctx, models.Transaction{
		TransactionID: req.TransactionID,
		UserID:        req.UserID,
		Currency:      updatedWallet.Currency,
		Amount:        req.Amount,
		Type:          "rollback",
		RequestID:     req.RequestID,
		BalanceAfter:  updatedWallet.Balance,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	})

	res := models.RollbackResponse{
		Balance:       updatedWallet.Balance,
		Status:        "OK",
		RequestID:     req.RequestID,
		TransactionID: req.TransactionID,
	}

	response.WriteJSONResponse(w, http.StatusOK, res)
}
