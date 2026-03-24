package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"CasinoLobbyBE/db"
	models "CasinoLobbyBE/types"
	"CasinoLobbyBE/utils/response"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Register handles POST /lobby/v1/register
func Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.GeneralErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	var missing []string
	if req.Email == "" {
		missing = append(missing, "email")
	}
	if req.Password == "" {
		missing = append(missing, "password")
	}
	if req.FirstName == "" {
		missing = append(missing, "first_name")
	}
	if len(missing) > 0 {
		response.GeneralErrorResponse(w, http.StatusBadRequest, "Missing required fields: "+strings.Join(missing, ", "))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Hash the password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		response.GeneralErrorResponse(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	// Generate a unique user ID
	userID := uuid.New().String()
	now := time.Now()

	userAuth := models.UserAuth{
		UserID:    userID,
		Email:     req.Email,
		Hash:      hashedPassword, // Store Bcrypt Hash
		CreatedAt: now,
		UpdatedAt: now,
	}

	userProfile := models.UserProfile{
		UserID:    userID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		CreatedAt: now,
		UpdatedAt: now,
	}

	wallet := models.Wallet{
		UserID:    userID,
		Balance:   "10000.00", // Assign a default mock balance
		Currency:  "USD",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Insert into DB
	_, err = db.GetCollection("userauth").InsertOne(ctx, userAuth)
	if err != nil {
		response.GeneralErrorResponse(w, http.StatusConflict, "Email already exists or DB error")
		return
	}

	_, err = db.GetCollection("userprofile").InsertOne(ctx, userProfile)
	if err != nil {
		response.GeneralErrorResponse(w, http.StatusInternalServerError, "Failed to create user profile")
		return
	}
	_, err = db.GetCollection("wallet").InsertOne(ctx, wallet)
	if err != nil {
		log.Println("❌ Wallet insert error:", err)
		response.GeneralErrorResponse(w, http.StatusInternalServerError, "Failed to initialize wallet")
		return
	}

	response.WriteJSONResponse(w, http.StatusCreated, map[string]string{"message": "User registered successfully"})
}

// Login handles POST /lobby/v1/login
func Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.GeneralErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	var missing []string
	if req.Email == "" {
		missing = append(missing, "email")
	}
	if req.Password == "" {
		missing = append(missing, "password")
	}
	if len(missing) > 0 {
		response.GeneralErrorResponse(w, http.StatusBadRequest, "Missing required fields: "+strings.Join(missing, ", "))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find user
	var userAuth models.UserAuth
	err := db.GetCollection("userauth").FindOne(ctx, bson.M{"email": req.Email}).Decode(&userAuth)
	if err == mongo.ErrNoDocuments {
		response.GeneralErrorResponse(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Verify password
	if !CheckPasswordHash(req.Password, userAuth.Hash) {
		response.GeneralErrorResponse(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Generate JWT Token
	token, err := GenerateToken(userAuth.UserID)
	if err != nil {
		response.GeneralErrorResponse(w, http.StatusInternalServerError, "Could not generate authentication token")
		return
	}

	// Set as an HttpOnly cookie for web client security
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	response.WriteJSONResponse(w, http.StatusOK, models.LoginResponse{Token: token, UserID: userAuth.UserID})
}

// Logout handles POST /lobby/v1/logout
func Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the auth cookie client-side by setting it in the past
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Path:     "/",
	})

	response.WriteJSONResponse(w, http.StatusOK, map[string]string{"message": "Successfully logged out"})
}
