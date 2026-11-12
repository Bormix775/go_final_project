package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// appPassword — пароль из окружения, считывается один раз в Init.
// Пустая строка означает, что аутентификация отключена.
var appPassword string

// InitAuth читает конфигурацию один раз при старте.
func InitAuth() {
	appPassword = os.Getenv("TODO_PASSWORD")
}

func passwordHash(pass string) string {
	h := sha256.Sum256([]byte(pass))
	return hex.EncodeToString(h[:])
}

func buildJWT(pass string) (string, error) {
	claims := jwt.MapClaims{
		"pass_hash": passwordHash(pass),
		"exp":       time.Now().Add(8 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(pass))
}

func validateJWT(jwtStr, pass string) bool {
	if jwtStr == "" {
		return false
	}
	token, err := jwt.Parse(jwtStr, func(t *jwt.Token) (any, error) {
		return []byte(pass), nil
	})
	if err != nil || !token.Valid {
		return false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}
	hash, ok := claims["pass_hash"].(string)
	if !ok {
		return false
	}
	return hash == passwordHash(pass)
}

func singInHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ошибка десериализации JSON"})
		return
	}
	if req.Password != appPassword {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Неверный пароль"})
		return
	}
	token, err := buildJWT(appPassword)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "не удалось создать токен"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if appPassword != "" {
			var jwtStr string
			if cookie, err := r.Cookie("token"); err == nil {
				jwtStr = cookie.Value
			}
			if !validateJWT(jwtStr, appPassword) {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}
