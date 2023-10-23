package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"go-mini-test/models"
	"go-mini-test/sqlstore"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("this-is-secret")

func createToken(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Hour * 12).Unix() // Token expiration time

	// Sign the token with a secret key
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func Login(w http.ResponseWriter, r *http.Request) {
	db := sqlstore.DbConn()
	reqBody, _ := io.ReadAll(r.Body)
	var userCred models.LoginRequest
	json.Unmarshal(reqBody, &userCred)

	username := userCred.Username
	password := userCred.Password

	var storedPassword string
	err := db.QueryRow("SELECT password FROM user_logins WHERE username = ?", username).Scan(&storedPassword)
	if err != nil {
		// fmt.Println(err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, errToken := createToken(username)
	if errToken != nil {
		http.Error(w, "Token creation failed", http.StatusUnauthorized)
		return
	}

	// Authentication successful
	json.NewEncoder(w).Encode(token)

}

func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode("No token provided")
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			// fmt.Println(err)
			json.NewEncoder(w).Encode("Token validation failed")
			return
		}

		if token.Valid {
			next(w, r)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode("Token is not valid")
		}
	}
}
