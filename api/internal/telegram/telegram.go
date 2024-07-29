package telegram

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"

	"github.com/Corray333/internship_app/internal/types"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Storage interface {
	UpdateUser(user *types.User) error
	CreateUser(user *types.User) error
	VerifyUser(chat_id int64) error
	GiveTasks(user_id int64, course_id string) error
	GetAllUsers() ([]types.User, error)
	GetTasks(user_id int64) ([]types.Task, error)
	GetUsersOnCourse(course_id string) ([]types.User, error)
	GetUserByID(user_id int64) (*types.User, error)
	GetCuratorOfUser(user_id int64) (int64, error)
	GetTask(user_id int64, task_id string) (*types.Task, error)
	GetCourse(course_id string) (*types.Course, error)
	IsCurator(uid int64) (bool, error)
	TaskDone(uid int64, task *types.Task) error
	RejectHomework(uid int64, taskID string) error
	SetUpdateData(updateID int, data string) error
	GetUpdateData(updateID int) (string, error)
}

const (
	StateWaitingFIO = iota + 1
	StateWaitingEmail
	StateWaitingDirection
	StateWaitingGroup
)

const (
	StateNothing = iota + 1
	StateWaitingUserTypePick
	StateWaitingMessageText
	StateWaitingMessageAttachment
	StateWaitingSending
)

type TelegramClient struct {
	bot    *tgbotapi.BotAPI
	admins map[int64]Admin
	store  Storage
}

type Admin struct {
	state int
	info  interface{}
}

var messages = []string{
	"Не понимаю, чего ты хочешь😅",
	"Прости, я тебя не понял🤔",
	"Что-то пошло не так, попробуй снова🙏",
	"Не могу разобраться, попробуй иначе😉",
	"Похоже, я тебя не понимаю😕",
	"Давай попробуем еще раз, я тебя не понял😊",
	"Может быть, я что-то упустил. Попробуй ещё раз😌",
	"Извини, я тебя не понял. Попробуй сформулировать иначе🤷‍♂️",
	"Я не совсем понял твоё действие. Попробуй что-то другое🙃",
	"Не могу распознать твой запрос. Попробуй снова🧐",
}

func NewClient(token string, store Storage) *TelegramClient {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("Failed to create bot: ", err)
	}

	bot.Debug = true

	return &TelegramClient{
		bot: bot,
		admins: map[int64]Admin{
			ANDREW_CHAT_ID: {
				state: StateNothing,
			},
			MARIA_CHAT_ID: {
				state: StateNothing,
			},
		},
		store: store,
	}
}

func (tg *TelegramClient) Run() {
	defer func() {
		if r := recover(); r != nil {
			tg.HandleError("panic: " + r.(string))
		}
	}()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tg.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		isCurator, err := tg.store.IsCurator(update.FromChat().ID)
		if err != nil {
			tg.HandleError("error while checking if user is curator: "+err.Error(), "update_id", update.UpdateID)
			continue
		}

		switch {
		case update.FromChat().ID == ANDREW_CHAT_ID:
			tg.handleAdminUpdate(update)
		case isCurator:
			tg.handleCuratorUpdate(update)
			continue
		case update.FromChat().ID == MARIA_CHAT_ID:
			if update.CallbackQuery != nil {
				tg.userAccepted(update)
			}
			continue
		default:
			tg.handleUserUpdate(update)
			continue
		}

	}
}

func (tg *TelegramClient) handleUserUpdate(update tgbotapi.Update) {

	if update.Message != nil && update.Message.Contact != nil {
		tg.handleContact(update)
		return
	}

	user, err := tg.store.GetUserByID(update.FromChat().ID)
	if err != nil {
		if err == sql.ErrNoRows {
			tg.sendWelcomeMessage(update.FromChat().ID)
			return
		}
		tg.HandleError("error while getting user from db: "+err.Error(), "update_id", update.UpdateID)
		return
	}

	fmt.Println()
	fmt.Printf("User: %+v\n", *user)
	fmt.Println()

	switch user.State {
	case StateWaitingFIO:
		tg.handleInputFIO(user, update)
	case StateWaitingEmail:
		tg.handleInputEmail(user, update)
	case StateWaitingDirection:
		tg.handleDirectionPick(user, update)
	case StateWaitingGroup:
		tg.groupJoined(user, update)
	default:
		msg := tgbotapi.NewMessage(update.FromChat().ID, messages[rand.Int()%len(messages)])
		if _, err := tg.bot.Send(msg); err != nil {
			tg.HandleError("error while sending message: "+err.Error(), "update_id", update.UpdateID)
			return
		}
	}

	if err := tg.store.UpdateUser(user); err != nil {
		tg.HandleError("error while updating user: "+err.Error(), "update_id", update.UpdateID)
	}
}

func (tg *TelegramClient) handleCuratorUpdate(update tgbotapi.Update) {
	if update.CallbackQuery == nil {
		return
	}

	data, err := tg.store.GetUpdateData(update.CallbackQuery.Message.MessageID)
	if err != nil {
		tg.HandleError("error while getting update data: "+err.Error(), "update_id", update.UpdateID)
		return
	}

	req := &HomeworkCheck{}
	if err := json.Unmarshal([]byte(data), req); err != nil {
		tg.HandleError("error while unmarshalling data: "+err.Error(), "update_id", update.UpdateID)
		return
	}

	task, err := tg.store.GetTask(req.UserID, req.TaskID)
	if err != nil {
		tg.HandleError("error while getting task: "+err.Error(), "update_id", update.UpdateID)
		return
	}

	if update.CallbackData() == HomeworkStatusApproved {
		if err := tg.store.TaskDone(req.UserID, task); err != nil {
			tg.HandleError("error while marking task as done: "+err.Error(), "update_id", update.UpdateID)
			return
		}
		msg := tgbotapi.NewMessage(req.UserID, "Домашняя работа одобрена, можешь переходить к следующему этапу😉")
		if _, err := tg.bot.Send(msg); err != nil {
			tg.HandleError("error while sending message: "+err.Error(), "update_id", update.UpdateID)
			return
		}
	} else if update.CallbackData() == HomeworkStatusRejected {
		if err := tg.store.RejectHomework(req.UserID, req.TaskID); err != nil {
			tg.HandleError("error while rejecting task: "+err.Error(), "update_id", update.UpdateID)
			return
		}
		msg := tgbotapi.NewMessage(req.UserID, "У тебя получилась хорошая работа, надо лишь немного довести ее до идеала) Прими правки и возвращайся с результатом😉")
		if _, err := tg.bot.Send(msg); err != nil {
			tg.HandleError("error while sending message: "+err.Error(), "update_id", update.UpdateID)
			return
		}
	} else {
		tg.HandleError("unknown status: "+update.CallbackData(), "update_id", update.UpdateID)
		return
	}

	cb := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
	if _, err := tg.bot.Request(cb); err != nil {
		tg.HandleError("error while sending callback: "+err.Error(), "update_id", update.UpdateID)
		return
	}

	del := tgbotapi.NewDeleteMessage(update.FromChat().ID, update.CallbackQuery.Message.MessageID)
	if _, err := tg.bot.Request(del); err != nil {
		tg.HandleError("error while deleting message: "+err.Error(), "update_id", update.UpdateID)
		return
	}
}
