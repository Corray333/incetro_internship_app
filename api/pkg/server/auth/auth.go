package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func CheckTelegramAuth(initData string) (int64, bool) {
	parsedData, _ := url.QueryUnescape(initData)
	chunks := strings.Split(parsedData, "&")
	var dataPairs [][]string
	hash := ""
	user := &struct {
		ID int64 `json:"id"`
	}{}

	// Filter and split the chunks
	for _, chunk := range chunks {
		if strings.HasPrefix(chunk, "user=") {
			parsedData = strings.TrimPrefix(chunk, "user=")
			if err := json.Unmarshal([]byte(parsedData), user); err != nil {
				slog.Error("Failed to unmarshal user data: " + err.Error())
				return 0, false
			}
		}
		if strings.HasPrefix(chunk, "hash=") {
			hash = strings.TrimPrefix(chunk, "hash=")
		} else {
			pair := strings.SplitN(chunk, "=", 2)
			dataPairs = append(dataPairs, pair)
		}
	}

	// Sort the data pairs by the key
	sort.Slice(dataPairs, func(i, j int) bool {
		return dataPairs[i][0] < dataPairs[j][0]
	})

	// Join the sorted data pairs into the initData string
	var sortedData []string
	for _, pair := range dataPairs {
		sortedData = append(sortedData, fmt.Sprintf("%s=%s", pair[0], pair[1]))
	}
	initData = strings.Join(sortedData, "\n")
	// Create the secret key using HMAC and the given token
	h := hmac.New(sha256.New, []byte("WebAppData"))
	h.Write([]byte(os.Getenv("BOT_TOKEN")))
	secretKey := h.Sum(nil)

	// Create the data check using the secret key and initData
	h = hmac.New(sha256.New, secretKey)
	h.Write([]byte(initData))
	dataCheck := h.Sum(nil)

	fmt.Println(initData)
	fmt.Printf("%x\n", dataCheck)
	fmt.Println(hash)

	return user.ID, fmt.Sprintf("%x", dataCheck) == hash
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
