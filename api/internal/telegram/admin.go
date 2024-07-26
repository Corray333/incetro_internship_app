package telegram

import (
	"encoding/json"
	"fmt"

	"github.com/Corray333/internship_app/internal/utils"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (tg *TelegramClient) handleAdminUpdate(update tgbotapi.Update) {
	admin, ok := tg.admins[update.FromChat().ID]
	if !ok {
		tg.admins[update.FromChat().ID] = Admin{state: StateNothing}
	}

	switch admin.state {
	case StateNothing:
		if update.Message != nil {
			if update.Message.IsCommand() && update.Message.Command() == "start" {
				msg := tgbotapi.NewMessage(update.FromChat().ID, "Приветствую тебя, всеотец. Для работы с ботом используй меню👇")
				msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton("Создать рассылку"),
					),
				)
				if _, err := tg.bot.Send(msg); err != nil {
					tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
				}
			} else if update.Message.Text == "Создать рассылку" {
				tg.admins[update.FromChat().ID] = Admin{state: StateWaitingUserTypePick}
				msg := tgbotapi.NewMessage(update.FromChat().ID, "Кому отправить сообщение?")
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("iOS", "iOS"),
						tgbotapi.NewInlineKeyboardButtonData("Flutter", "flutter"),
						tgbotapi.NewInlineKeyboardButtonData("Android", "android"),
					),
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Всем пользователям", "all"),
					),
				)
				if _, err := tg.bot.Send(msg); err != nil {
					tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
				}
			}
		}
	case StateWaitingUserTypePick:
		if update.CallbackQuery != nil {
			cb := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
			if _, err := tg.bot.Send(cb); err != nil {
				tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
			}

			del := tgbotapi.NewDeleteMessage(update.FromChat().ID, update.CallbackQuery.Message.MessageID)
			if _, err := tg.bot.Send(del); err != nil {
				tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
			}

			info := AdminRequestSending{}
			info.UserType = update.CallbackData()
			admin.info = info
			tg.admins[update.FromChat().ID] = Admin{state: StateWaitingMessageText, info: info}
			msg := tgbotapi.NewMessage(update.FromChat().ID, "Теперь отправь сообщение, которое нужно переслать всем пользователям👇")
			if _, err := tg.bot.Send(msg); err != nil {
				tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
			}
		}
	case StateWaitingMessageText:
		if update.Message != nil {

			tg.admins[update.FromChat().ID] = Admin{
				state: StateNothing,
			}

			info, ok := admin.info.(AdminRequestSending)
			if !ok {
				tg.HandleError("error while getting admin info: wrong message type", "update_id", update.UpdateID)
			}
			info.Message = tgbotapi.MessageConfig{
				Text:     update.Message.Text,
				Entities: update.Message.Entities,
			}
			admin.info = info
			admin.state = StateWaitingMessageAttachment
			tg.admins[update.FromChat().ID] = admin

			msg := tgbotapi.NewMessage(update.FromChat().ID, "Теперь добавь вложения, если они нужны. Это могут быть картинки, гифка или файлы. Когда отправишь все вложения, напиши в чате 'стоп'")
			if _, err := tg.bot.Send(msg); err != nil {
				tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
			}
		}
	case StateWaitingMessageAttachment:
		info, ok := admin.info.(AdminRequestSending)
		if !ok {
			tg.HandleError("error while getting admin info: wrong message type", "update_id", update.UpdateID)
			msg := tgbotapi.NewMessage(update.FromChat().ID, "Сорри, что-то пошло не так😬 Давай попробуем сначала.")
			if _, err := tg.bot.Send(msg); err != nil {
				tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
				return
			}
			tg.admins[update.FromChat().ID] = Admin{state: StateNothing}
		}
		switch {
		case update.Message.Photo != nil:
			info.AttachmentType = AttachmentImage
			info.Attachments = append(info.Attachments, tgbotapi.FileID(update.Message.Photo[0].FileID))
			admin.info = info
			tg.admins[update.FromChat().ID] = admin

		case update.Message.Document != nil:
			info.AttachmentType = AttachmentFile
			info.Attachments = append(info.Attachments, tgbotapi.FileID(update.Message.Document.FileID))
			admin.info = info
			tg.admins[update.FromChat().ID] = admin

		case update.Message.Animation != nil:
			info.AttachmentType = AttachmentAnimation
			info.Attachments = append(info.Attachments, tgbotapi.FileID(update.Message.Animation.FileID))
			msg := tgbotapi.NewAnimation(update.FromChat().ID, info.Attachments[0])
			msg.CaptionEntities = info.Message.Entities
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Отправить", "send"),
					tgbotapi.NewInlineKeyboardButtonData("Отмена", "cancel"),
				),
			)
			if _, err := tg.bot.Send(msg); err != nil {
				tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
				return
			}
			admin.info = info
			tg.admins[update.FromChat().ID] = admin
		case update.Message.Text == "стоп":
			admin.state = StateWaitingSending
			tg.admins[update.FromChat().ID] = admin
			switch info.AttachmentType {
			case AttachmentImage:
				if len(info.Attachments) == 1 {
					msg := tgbotapi.NewPhoto(update.FromChat().ID, info.Attachments[0])
					msg.CaptionEntities = info.Message.Entities
					msg.Caption = info.Message.Text
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Отправить", "send"),
							tgbotapi.NewInlineKeyboardButtonData("Отмена", "cancel"),
						),
					)
					if _, err := tg.bot.Send(msg); err != nil {
						tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
						return
					}

				} else {
					mediaGroup := make([]interface{}, len(info.Attachments))
					for i := range info.Attachments {
						mediaGroup[i] = tgbotapi.NewInputMediaPhoto(info.Attachments[i])
					}
					mg := tgbotapi.NewMediaGroup(update.FromChat().ID, mediaGroup)
					if _, err := tg.bot.Send(mg); err != nil {
						tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
						return
					}
					msg := tgbotapi.NewMessage(update.FromChat().ID, info.Message.Text)
					msg.Entities = info.Message.Entities
					msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
						tgbotapi.NewInlineKeyboardRow(
							tgbotapi.NewInlineKeyboardButtonData("Отправить", "send"),
							tgbotapi.NewInlineKeyboardButtonData("Отмена", "cancel"),
						),
					)
					if _, err := tg.bot.Send(msg); err != nil {
						tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
						return
					}
				}
			case AttachmentFile:
				for _, file := range info.Attachments {
					msg := tgbotapi.NewDocument(update.FromChat().ID, file)
					if _, err := tg.bot.Send(msg); err != nil {
						tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
						return
					}
				}
				msg := tgbotapi.NewMessage(update.FromChat().ID, info.Message.Text)
				msg.Entities = info.Message.Entities
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Отправить", "send"),
						tgbotapi.NewInlineKeyboardButtonData("Отмена", "cancel"),
					),
				)
				if _, err := tg.bot.Send(msg); err != nil {
					tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
					return
				}
			}
		}
	case StateWaitingSending:
		cb := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		if _, err := tg.bot.Send(cb); err != nil {
			tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
			return
		}

		tg.admins[update.FromChat().ID] = Admin{state: StateNothing}

		del := tgbotapi.NewDeleteMessage(update.FromChat().ID, update.CallbackQuery.Message.MessageID)
		if _, err := tg.bot.Send(del); err != nil {
			tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
			return
		}

		info, ok := admin.info.(AdminRequestSending)
		if !ok {
			tg.HandleError("error while sending status: wrong info type", "update_id", update.UpdateID)
		}

		if update.CallbackQuery == nil {
			tg.HandleError("error while sending status: callback query is nil", "update_id", update.UpdateID)
		}

		if update.CallbackQuery.Data == "cancel" {
			msg := tgbotapi.NewMessage(update.FromChat().ID, "Отмена отправки сообщения")
			if _, err := tg.bot.Send(msg); err != nil {
				tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
				return
			}
			tg.admins[update.FromChat().ID] = Admin{state: StateNothing}
			return
		}

		switch info.AttachmentType {
		case AttachmentImage:
			users, err := tg.store.GetUsersOnCourse(info.UserType)
			if err != nil {
				tg.HandleError("error while getting users on course: "+err.Error(), "update_id", update.UpdateID)
			}
			for _, user := range users {
				if len(info.Attachments) == 1 {
					msg := tgbotapi.NewPhoto(user.UserID, info.Attachments[0])
					msg.CaptionEntities = info.Message.Entities
					msg.Caption = info.Message.Text
					if _, err := tg.bot.Send(msg); err != nil {
						tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
						return
					}
				} else {
					mediaGroup := make([]interface{}, len(info.Attachments))
					for i := range info.Attachments {
						mediaGroup[i] = tgbotapi.NewInputMediaPhoto(info.Attachments[i])
					}
					mg := tgbotapi.NewMediaGroup(user.UserID, mediaGroup)
					if _, err := tg.bot.Send(mg); err != nil {
						tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
						return
					}
					msg := tgbotapi.NewMessage(user.UserID, info.Message.Text)
					msg.Entities = info.Message.Entities
					if _, err := tg.bot.Send(msg); err != nil {
						tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
						return
					}
				}
			}

		case AttachmentFile:
			users, err := tg.store.GetUsersOnCourse(info.UserType)
			if err != nil {
				tg.HandleError("error while getting users on course: "+err.Error(), "update_id", update.UpdateID)
			}
			for _, user := range users {
				for _, file := range info.Attachments {
					msg := tgbotapi.NewDocument(user.UserID, file)
					if _, err := tg.bot.Send(msg); err != nil {
						tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
						return
					}
				}
				msg := tgbotapi.NewMessage(user.UserID, info.Message.Text)
				msg.Entities = info.Message.Entities
				msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Отправить", "send"),
						tgbotapi.NewInlineKeyboardButtonData("Отмена", "cancel"),
					),
				)
				if _, err := tg.bot.Send(msg); err != nil {
					tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
					return
				}
			}

		case AttachmentAnimation:
			users, err := tg.store.GetUsersOnCourse(info.UserType)
			if err != nil {
				tg.HandleError("error while getting users on course: "+err.Error(), "update_id", update.UpdateID)
			}
			for _, user := range users {
				msg := tgbotapi.NewAnimation(user.UserID, info.Attachments[0])
				msg.CaptionEntities = info.Message.Entities
				if _, err := tg.bot.Send(msg); err != nil {
					tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
					return
				}
			}
		}
		msg := tgbotapi.NewMessage(update.FromChat().ID, "Рассылка запущена😉")
		if _, err := tg.bot.Send(msg); err != nil {
			tg.HandleError("error sending message: "+err.Error(), "update id", update.UpdateID)
			return
		}

	}

}

const (
	AttachmentFile = iota + 1
	AttachmentImage
	AttachmentAnimation
)

type AdminRequestSending struct {
	UserType       string
	Message        tgbotapi.MessageConfig
	AttachmentType int
	Attachments    []tgbotapi.RequestFileData
}

func (tg *TelegramClient) SendHomework(uid int64, taskID string, message string) error {
	user, err := tg.store.GetUserByID(uid)
	if err != nil {
		return err
	}

	curatorID, err := tg.store.GetCuratorOfUser(uid)
	if err != nil {
		return err
	}

	task, err := tg.store.GetTask(uid, taskID)
	if err != nil {
		return err
	}

	text := fmt.Sprintf("**Пользователь** [%s](%s) **отправил домашнюю работу к задаче %s:\n\n%s", user.FIO, "t.me/"+user.Username, task.Title, utils.EscapeMarkdownV2(message))

	data, err := json.Marshal(HomeworkCheck{
		TaskID: taskID,
		UserID: uid,
	})
	if err != nil {
		return err
	}

	msg := tgbotapi.NewMessage(curatorID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Принять", HomeworkStatusApproved),
			tgbotapi.NewInlineKeyboardButtonData("Отклонить", HomeworkStatusRejected),
		),
	)
	sent, err := tg.bot.Send(msg)
	if err != nil {
		return err
	}

	if err := tg.store.SetUpdateData(sent.MessageID, string(data)); err != nil {
		return err
	}

	return nil
}

type HomeworkCheck struct {
	TaskID string `json:"task_id"`
	UserID int64  `json:"user_id"`
}

var (
	HomeworkStatusApproved = "approved"
	HomeworkStatusRejected = "rejected"
)
