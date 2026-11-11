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
	return claims["pass_hash"] == passwordHash(pass)
}

func singInHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, map[string]string{"error": "ошибка десериализации JSON"})
		return
	}
	pass := os.Getenv("TODO_PASSWORD")
	if req.Password != pass {
		writeJSON(w, map[string]string{"error": "Неверный пароль"})
		return
	}
	token, err := buildJWT(pass)
	if err != nil {
		writeJSON(w, map[string]string{"error": "не удалось создать токен"})
		return
	}
	writeJSON(w, map[string]string{"token": token})
}

func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) > 0 {
			var jwtStr string
			if cookie, err := r.Cookie("token"); err == nil {
				jwtStr = cookie.Value
			}
			if !validateJWT(jwtStr, pass) {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}
