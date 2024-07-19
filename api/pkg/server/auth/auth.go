package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func СheckTelegramAuth(data map[string]string) bool {
	// Сортировка ключей для формирования строки для проверки хэша
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Формирование строки для проверки хэша
	var dataCheckStrings []string
	for _, k := range keys {
		if k != "hash" { // исключаем поле 'hash' из строки для проверки хэша
			dataCheckStrings = append(dataCheckStrings, fmt.Sprintf("%s=%s", k, data[k]))
		}
	}
	dataCheckString := strings.Join(dataCheckStrings, "\n")

	// Вычисление HMAC-SHA256 хэша
	key := []byte(os.Getenv("BOT_TOKEN"))
	h := hmac.New(sha256.New, key)
	h.Write([]byte(dataCheckString))
	calculatedHash := hex.EncodeToString(h.Sum(nil))

	// Сравнение вычисленного хэша с переданным хэшем
	return calculatedHash == data["hash"]
}

const (
	AccessTokenLifeTime  = time.Minute * 60
	RefreshTokenLifeTime = time.Hour * 24 * 7
)

var secretKey []byte

// init initializes the secret key from the environment variable
func init() {
	secretKey = []byte(os.Getenv("SECRET_KEY"))
}

func NewAuthMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		slog.Info("auth middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			creds, err := VerifyToken(r.Header.Get("Authorization"))
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				slog.Error("Unauthorized: " + err.Error())
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), "creds", creds))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// Hash hashes the password using bcrypt package
func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// Verify checks if the hashed password is equal to the password using bcrypt package
func Verify(hashed, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	return err == nil
}

// CreateToken creates a new JWT token by the email
func CreateToken(id int64, lifeTime time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"id":  id,
			"exp": time.Now().Add(lifeTime).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// VerifyToken checks if the JWT is valid
func VerifyToken(tokenString string) (Credentials, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return Credentials{}, err
	}

	if !token.Valid {
		return Credentials{}, fmt.Errorf("invalid token")
	}

	exp, err := token.Claims.GetExpirationTime()
	if err != nil {
		return Credentials{}, err
	}
	creds := Credentials{
		ID:  int64(claims["id"].(float64)),
		Exp: exp.Time,
	}

	return creds, nil
}

type Credentials struct {
	ID  int64 `json:"id,omitempty" db:"user_id"`
	Exp time.Time
}

func ExtractCredentials(tokenString string) (*Credentials, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	exp, err := token.Claims.GetExpirationTime()
	if err != nil {
		return nil, err
	}
	credentials := Credentials{
		ID:  int64(claims["id"].(float64)),
		Exp: exp.Time,
	}
	return &credentials, nil
}

type Storage interface {
	CheckAndUpdateRefresh(id int64, refresh string) (string, error)
}

func RefreshAccessToken(store Storage, refresh string) (string, string, error) {

	token, err := jwt.Parse(refresh, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return "", "", err
	}

	if !token.Valid {
		return "", "", fmt.Errorf("invalid refresh token")
	}

	creds, err := ExtractCredentials(refresh)
	if err != nil {
		return "", "", err
	}

	newRefresh, err := store.CheckAndUpdateRefresh(creds.ID, refresh)
	if err != nil {
		return "", "", err
	}
	newAccess, err := CreateToken(creds.ID, AccessTokenLifeTime)
	if err != nil {
		return "", "", err
	}
	return newAccess, newRefresh, nil

}
