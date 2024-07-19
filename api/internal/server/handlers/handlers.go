package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Corray333/internship_app/internal/telegram"
	"github.com/Corray333/internship_app/internal/types"
	"github.com/Corray333/internship_app/internal/utils"
	"github.com/Corray333/internship_app/pkg/server/auth"
	"github.com/go-chi/chi/v5"
)

type Storage interface {
	GetTasks(user_id int64) ([]types.Task, error)
	GetTask(user_id int64, task_id string) (*types.Task, error)
	SaveHomework(user_id int64, taskID string, homework string) error
	SetRefresh(uid int64, refresh string) error
	RefreshToken(id int64, oldRefresh string) (string, string, error)
	TaskDone(uid int64, task *types.Task) error
}

type user struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  string `json:"auth_date"`
	Hash      string `json:"hash"`
}

type ListTasksResponse []types.Task

// @Summary Получить список задач
// @Description Получить список задач, доступных пользователю
// @Tags tasks
// @Accept json
// @Produce json
// @Param Authorization header string true "Access JWT"
// @Success 200 {array} ListTasksResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/tasks [get]
func ListTasks(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		creds := r.Context().Value("creds").(auth.Credentials)
		if creds.ID == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		uid := creds.ID

		tasks, err := store.GetTasks(uid)
		if err != nil {
			slog.Error("error while getting tasks: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tasks, err = utils.TopologicalSort(tasks)
		if err != nil {
			slog.Error("error while sorting tasks: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(tasks); err != nil {
			slog.Error("error while encoding tasks: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

type GetTaskResponse types.Task

// @Summary Получить задачу. Forbidden if status is Not started
// @Description Получить задачу (если она начата)
// @Tags tasks
// @Accept json
// @Produce json
// @Param Authorization header string true "Access JWT"
// @Param task_id path string true "Task ID"
// @Success 200 {object} GetTaskResponse
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/tasks/{task_id} [get]
func GetTask(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID := chi.URLParam(r, "task_id")
		if taskID == "" {
			slog.Error("error while getting task id")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		creds := r.Context().Value("creds").(auth.Credentials)
		if creds.ID == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		uid := creds.ID

		task, err := store.GetTask(uid, taskID)
		if err != nil {
			slog.Error("error while getting task: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if task.Status == types.TaskStatusNotStarted {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(task); err != nil {
			slog.Error("error while encoding task: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}
}

type SaveHomeworkRequest struct {
	Homework string `json:"homework"`
}

// @Summary Сдать домашнюю работу
// @Description Сдать домашнюю работу для определенной задачи
// @Tags tasks
// @Accept json
// @Produce json
// @Param Authorization header string true "Access JWT"
// @Param task_id path string true "Task ID"
// @Param request body SaveHomeworkRequest true "Homework Data"
// @Success 201 {string} string "Created"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/tasks/{task_id}/homework [post]
// @Router /api/tasks/{task_id}/homework [patch]
func SaveHomework(tg *telegram.TelegramClient, store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		creds := r.Context().Value("creds").(auth.Credentials)
		if creds.ID == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		taskID := chi.URLParam(r, "task_id")
		if taskID == "" {
			slog.Error("error while getting task id")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		req := &SaveHomeworkRequest{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			slog.Error("error while decoding save homework request: " + err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		uid := creds.ID

		if err := store.SaveHomework(uid, taskID, req.Homework); err != nil {
			slog.Error("error while saving homework: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tg.SendHomework(uid, taskID, req.Homework); err != nil {
			slog.Error("error while sending homework: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			tg.HandleError("error while sending homework: "+err.Error(), "user_id", uid, "task_id", taskID)
			return
		}

		w.WriteHeader(http.StatusCreated)

	}
}

type LoginRequest struct {
	User user `json:"user"`
}

// @Summary Вход
// @Description Вход в профиль и получение токенов
// @Tags users
// @Accept json
// @Produce json
// @Param request body LoginRequest true "User Info"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/users/login [post]
func Login(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &struct {
			User map[string]string `json:"user"`
		}{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			slog.Error("error while decoding save homework request: " + err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// if allowed := auth.СheckTelegramAuth(req.User); !allowed {
		// 	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		// 	return
		// }

		uidRaw, ok := req.User["id"]
		if !ok {
			slog.Error("error while getting user id")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		uid, err := strconv.Atoi(uidRaw)
		if err != nil {
			slog.Error("error while converting user id to int")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		refresh, err := auth.CreateToken(int64(uid), auth.RefreshTokenLifeTime)
		if err != nil {
			http.Error(w, "Failed to create token", http.StatusInternalServerError)
			slog.Error("Failed to create token: " + err.Error())
			return
		}

		if err := store.SetRefresh(int64(uid), refresh); err != nil {
			http.Error(w, "Failed to set refresh token", http.StatusInternalServerError)
			slog.Error("Failed to set refresh token: " + err.Error())
			return
		}

		fmt.Println()
		fmt.Println("Login refresh: ", refresh)
		fmt.Println()

		creds, err := auth.ExtractCredentials(refresh)
		if err != nil {
			http.Error(w, "Failed to insert user", http.StatusInternalServerError)
			slog.Error("Failed to insert user: " + err.Error())
			return
		}

		cookie := http.Cookie{
			Name:     "Refresh",
			Value:    refresh,
			Expires:  creds.Exp,
			HttpOnly: true,
			Path:     "/",
			SameSite: http.SameSiteNoneMode,
			Secure:   true,
		}

		http.SetCookie(w, &cookie)

		token, err := auth.CreateToken(int64(uid), auth.AccessTokenLifeTime)
		if err != nil {
			http.Error(w, "Failed to create token", http.StatusInternalServerError)
			slog.Error("Failed to create token: " + err.Error())
			return
		}

		w.Header().Set("Authorization", token)

		w.WriteHeader(http.StatusOK)
	}
}

// @Summary Обновить токена
// @Description Обновить access и refresh токены
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/users/refresh-tokens [patch]
func RefreshTokens(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		refreshCookie, err := r.Cookie("Refresh")
		if err != nil {
			http.Error(w, "Failed to get refresh cookie", http.StatusUnauthorized)
			slog.Error("Failed to get refresh cookie: " + err.Error())
			return
		}
		if refreshCookie.Value == "" {
			http.Error(w, "Failed to get refresh cookie", http.StatusUnauthorized)
			slog.Error("Failed to get refresh cookie")
			return
		}

		creds, err := auth.ExtractCredentials(refreshCookie.Value)
		if err != nil {
			http.Error(w, "Failed to extract credentials", http.StatusBadRequest)
			slog.Error("Failed to extract credentials: " + err.Error())
			return
		}
		access, refresh, err := store.RefreshToken(creds.ID, refreshCookie.Value)
		if err != nil {
			http.Error(w, "Failed to refresh token", http.StatusInternalServerError)
			slog.Error("Failed to refresh token: " + err.Error())
			return
		}

		creds, err = auth.ExtractCredentials(refresh)
		if err != nil {
			http.Error(w, "Failed to insert user", http.StatusInternalServerError)
			slog.Error("Failed to insert user: " + err.Error())
			return
		}

		cookie := http.Cookie{
			Name:     "Refresh",
			Value:    refresh,
			Expires:  creds.Exp,
			HttpOnly: true,
			Path:     "/",
			SameSite: http.SameSiteNoneMode,
			Secure:   true,
		}

		http.SetCookie(w, &cookie)

		w.Header().Set("Authorization", access)
		w.WriteHeader(http.StatusOK)
	}
}

// @Summary Задача выполнена
// @Description Отметить теоретическую задачу выполненной, следующая задача получает статус "В процессе"
// @Tags tasks
// @Accept json
// @Produce json
// @Param Authorization header string true "Access JWT"
// @Param task_id path string true "Task ID"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/tasks/{task_id} [patch]
func TaskDone(store Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		taskID := chi.URLParam(r, "task_id")
		if taskID == "" {
			slog.Error("error while getting task id")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		creds := r.Context().Value("creds").(auth.Credentials)
		if creds.ID == 0 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		uid := creds.ID

		task, err := store.GetTask(uid, taskID)
		if err != nil {
			slog.Error("error while getting task: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if task.Status != types.TaskStatusInProgress {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if task.Type == "Практика" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		task.CompletedAt = time.Now().Unix()
		task.Status = types.TaskStatusDone

		if err := store.TaskDone(uid, task); err != nil {
			slog.Error("error while updating task: " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

	}
}
