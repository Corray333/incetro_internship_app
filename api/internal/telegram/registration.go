package telegram

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/mail"
	"os"
	"regexp"
	"strconv"

	"github.com/Corray333/internship_app/internal/notion"
	"github.com/Corray333/internship_app/internal/types"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (tg *TelegramClient) sendWelcomeMessage(chatID int64) {
	_, err := tg.store.GetUserByID(chatID)
	if err != nil {
		if err == sql.ErrNoRows {
			button := tgbotapi.NewKeyboardButtonContact("Отправить контакт")
			keyboard := tgbotapi.NewReplyKeyboard([]tgbotapi.KeyboardButton{button})
			msg := tgbotapi.NewMessage(chatID, "Привет! Это команда Incetro.\nЧтобы зарегистрироваться на стажировку, поделись своим контактом🤙")
			msg.ReplyMarkup = keyboard
			if _, err := tg.bot.Send(msg); err != nil {
				tg.HandleError("error while sending message: "+err.Error(), "chat_id", chatID)
				return
			}
			return
		}
		tg.HandleError("error while getting user from db: "+err.Error(), "chat_id", chatID)
		return
	}
	msg := tgbotapi.NewMessage(chatID, "Прости, но я не понимаю, что ты от меня хочешь😥")
	if _, err := tg.bot.Send(msg); err != nil {
		tg.HandleError("error while sending message: "+err.Error(), "chat_id", chatID)
		return
	}

}

func (tg *TelegramClient) handleInputFIO(user *types.User, update tgbotapi.Update) {
	re := regexp.MustCompile(`^([A-Za-zА-Яа-яЁё]+[ \t]*)+$`)
	if !re.MatchString(update.Message.Text) {
		var msg tgbotapi.Chattable
		switch user.Fails {
		case 0:
			msg = tgbotapi.NewMessage(update.FromChat().ID, "Пожалуйста, введи только имя и фамилию)")
		case 1:
			msg = tgbotapi.NewMessage(update.FromChat().ID, "Прошууу, введи только имя и фамилию🙄")
		case 2:
			msg = tgbotapi.NewMessage(update.FromChat().ID, "Ну тебе что, сложно ввести только имя и фамилию?💩")
		case 3:
			msg = tgbotapi.NewMessage(update.FromChat().ID, "Давай, я верю в тебя! Только имя и фамилия! 💪")
		case 4:
			msg = tgbotapi.NewMessage(update.FromChat().ID, "Серьёзно? Имя и фамилия, пожалуйста! 🙃")
		case 5:
			msg = tgbotapi.NewMessage(update.FromChat().ID, "Нуууу, ты можешь! Просто имя и фамилия! 😤")
		case 6:
			msg = tgbotapi.NewMessage(update.FromChat().ID, "Ты точно умеешь читать, да? Имя. Фамилия. 🧐")
		case 7:
			msg = tgbotapi.NewMessage(update.FromChat().ID, "А может, тебе помочь? Имя и фамилия, давай! 😅")
		case 8:
			msg = tgbotapi.NewMessage(update.FromChat().ID, "Ладно, это уже не смешно. Имя и фамилия, ок? 😑")
		case 9:
			msg = tgbotapi.NewMessage(update.FromChat().ID, "Последний раз повторяю - имя и фамилия!😡")
		default:
			anim := tgbotapi.NewAnimation(update.FromChat().ID, tgbotapi.FileURL("https://i.giphy.com/media/v1.Y2lkPTc5MGI3NjExNnBqcXV2eDYxeG9xcjgweDh1dms5dnExdjIzbndpYTdzZzY5MHlwaiZlcD12MV9pbnRlcm5hbF9naWZfYnlfaWQmY3Q9Zw/xT1R9Y1cmnniQDyL5K/giphy.gif"))
			anim.Caption = "Ой, всё, наши индусы устали придумывать ответы на твои косяки, просто введи имя и фамилию😫"
			msg = anim
		}
		if _, err := tg.bot.Send(msg); err != nil {
			tg.HandleError("error while sending message: "+err.Error(), "chat_id", update.FromChat().ID)
			return
		}
		user.Fails++
		return
	}
	user.FIO = update.Message.Text
	user.State = StateWaitingEmail
	user.Fails = 0

	msg := tgbotapi.NewMessage(update.FromChat().ID, "Осталось совсем немного 🌝 отправишь свою рабочую / контактную почту?)")
	if _, err := tg.bot.Send(msg); err != nil {
		tg.HandleError("error while sending message: "+err.Error(), "chat_id", update.FromChat().ID)
		return
	}

}

func (tg *TelegramClient) handleInputEmail(user *types.User, update tgbotapi.Update) {
	if _, err := mail.ParseAddress(update.Message.Text); err != nil {
		var msg tgbotapi.Chattable
		switch user.Fails {
		case 0:
			msg = tgbotapi.NewMessage(update.FromChat().ID, "Введи настоящую электронную почту)")
		case 1:
			msg = tgbotapi.NewMessage(update.FromChat().ID, "Ну введи электронную почту🥺")
		case 2:
			msg = tgbotapi.NewMessage(update.FromChat().ID, "Пожалуйста, введи настоящую электронную почту! 📧")
		case 3:
			msg = tgbotapi.NewMessage(update.FromChat().ID, "Эй, мне нужна твоя настоящая электронная почта! 🙄")
		case 4:
			msg = tgbotapi.NewMessage(update.FromChat().ID, "Давай, без фокусов. Электронная почта, пожалуйста! 😅")
		case 5:
			msg = tgbotapi.NewMessage(update.FromChat().ID, "Ты сможешь! Введи правильную электронную почту! 💪")
		case 6:
			msg = tgbotapi.NewMessage(update.FromChat().ID, "Ну это уже смешно. Настоящую почту, пожалуйста. 🧐")
		case 7:
			msg = tgbotapi.NewMessage(update.FromChat().ID, "Электронную почту в студию! 🎤")
		case 8:
			msg = tgbotapi.NewMessage(update.FromChat().ID, "Серьёзно, без правильной почты никак. 🙃")
		case 9:
			msg = tgbotapi.NewMessage(update.FromChat().ID, "Последний раз! Введи свою настоящую почту! 😡")
		default:
			anim := tgbotapi.NewAnimation(update.FromChat().ID, tgbotapi.FileURL("https://i.giphy.com/media/v1.Y2lkPTc5MGI3NjExN3l1M2VnbWVqbjVxamVxMGc4dGMxempuMGh1bmYwM3VxdTQ3NXZleSZlcD12MV9pbnRlcm5hbF9naWZfYnlfaWQmY3Q9Zw/l4HnKwiJJaJQB04Zq/giphy.gif"))
			anim.Caption = "Сорри, у нашего разработчика кончилось воображение, теперь сообщения будут повторяться("
			msg = anim
		}
		if _, err := tg.bot.Send(msg); err != nil {
			tg.HandleError("error while sending message: "+err.Error(), "chat_id", update.FromChat().ID)
			return
		}
		user.Fails++
		return
	}
	user.Email = update.Message.Text
	user.State = StateWaitingDirection
	user.Fails = 0

	courses, err := notion.GetCourses()
	if err != nil {
		msg := tgbotapi.NewMessage(user.UserID, "Упс, у нас возникли технические проблемы, попробуй зайти позже😬")
		if _, err := tg.bot.Send(msg); err != nil {
			tg.HandleError("error while sending message: "+err.Error(), "chat_id", update.FromChat().ID)
			return
		}
		slog.Error("Error getting courses:" + err.Error())
		return
	}

	buttons := [][]tgbotapi.InlineKeyboardButton{}
	for _, course := range courses {
		buttons = append(buttons, []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(course.ShortName, course.NotionID)})
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
	msg := tgbotapi.NewMessage(update.FromChat().ID, "Выбери курс:")
	msg.ReplyMarkup = keyboard
	if _, err := tg.bot.Send(msg); err != nil {
		tg.HandleError("error while sending message: "+err.Error(), "chat_id", update.FromChat().ID)
		return
	}
}

func (tg *TelegramClient) handleDirectionPick(user *types.User, update tgbotapi.Update) {
	callbackQuery := update.CallbackQuery

	selectedCourse := callbackQuery.Data

	if user.State != StateWaitingDirection {
		return
	}
	user.Course = &selectedCourse
	user.State = StateWaitingGroup

	callback := tgbotapi.NewCallback(callbackQuery.ID, "Спасибо за выбор курса")
	if _, err := tg.bot.Request(callback); err != nil {
		tg.HandleError("error sending callback: "+err.Error(), "update_id", update.UpdateID)
		return
	}

	deleteMsg := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
	if _, err := tg.bot.Request(deleteMsg); err != nil {
		tg.HandleError("error deleting message: "+err.Error(), "update_id", update.UpdateID)
		return
	}

	course, err := tg.store.GetCourse(*user.Course)
	if err != nil {
		slog.Error("Error getting course:" + err.Error())
		tg.HandleError("error getting course: "+err.Error(), "update_id", update.UpdateID)
		return
	}

	msg := tgbotapi.NewMessage(callbackQuery.From.ID, "Теперь вступи в группу твоего курса, в ней ты найдешь много полезной информации и сможешь общаться с другими стажерами.")
	keyboard := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonURL("Вступить в группу", course.Invite),
	}, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Я вступил в групу", "joined_group"),
	})
	msg.ReplyMarkup = keyboard
	if _, err := tg.bot.Request(msg); err != nil {
		tg.HandleError("error sending message: "+err.Error(), "update_id", update.UpdateID)
		return
	}

}

func (tg *TelegramClient) groupJoined(user *types.User, update tgbotapi.Update) {
	callbackQuery := update.CallbackQuery
	fmt.Println()
	fmt.Printf("User check2: %+v\n", user)
	fmt.Println(*user.Course)
	fmt.Println()

	course, err := tg.store.GetCourse(*user.Course)

	if err != nil {
		tg.HandleError("error while getting course: "+err.Error(), "update_id", update.UpdateID)
		return
	}

	chatID := course.GroupID
	userID := callbackQuery.From.ID

	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		slog.Error(err.Error())
	}

	chatMember, err := bot.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: chatID,
			UserID: userID,
		},
	})
	if err != nil {
		slog.Error(err.Error())
	}

	// Check the status of the user in the chat
	if chatMember.Status != "creator" && chatMember.Status != "administrator" && chatMember.Status != "member" {
		callback := tgbotapi.NewCallback(callbackQuery.ID, "Вступи в группу")
		if _, err := tg.bot.Request(callback); err != nil {
			tg.HandleError("error sending callback: "+err.Error(), "update_id", update.UpdateID)
			return
		}
		msg := tgbotapi.NewMessage(userID, "Вступи в группу😑")
		if _, err := tg.bot.Request(msg); err != nil {
			tg.HandleError("error sending message: "+err.Error(), "update_id", update.UpdateID)
			return
		}
		return
	}

	callback := tgbotapi.NewCallback(callbackQuery.ID, "Спасибо")
	if _, err := tg.bot.Request(callback); err != nil {
		tg.HandleError("error sending callback: "+err.Error(), "update_id", update.UpdateID)
		return
	}

	deleteMsg := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
	if _, err := tg.bot.Request(deleteMsg); err != nil {
		tg.HandleError("error deleting message: "+err.Error(), "update_id", update.UpdateID)
		return
	}

	msg := tgbotapi.NewMessage(callbackQuery.From.ID, "Спасибо, твоя заявка принята, скоро мы вышлем инструкции для прохождения стажировки")
	if _, err := tg.bot.Request(msg); err != nil {
		tg.HandleError("error sending message: "+err.Error(), "update_id", update.UpdateID)
		return
	}

	tg.createUser(callbackQuery.Message.Chat.ID, user)
}

func (tg *TelegramClient) createUser(chat_id int64, user *types.User) {
	userID, err := notion.CreateUser(chat_id, user)
	if err != nil {
		msg := tgbotapi.NewMessage(chat_id, "Ошибка при создании пользователя.")
		if _, err := tg.bot.Request(msg); err != nil {
			tg.HandleError("error sending message: "+err.Error(), "user", user)
			return
		}
		log.Println("Error creating user:", err)
		return
	}
	user.ProfileID = &userID

	msg := tgbotapi.NewMessage(MARIA_CHAT_ID, fmt.Sprintf("Зарегистрирован новый пользователь: [%s](%s)", user.Username, "https://t.me/"+user.Username))
	button := tgbotapi.NewInlineKeyboardButtonData("Пользователь оформлен", strconv.Itoa(int(chat_id)))
	keyboard := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{button})
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard
	if _, err := tg.bot.Request(msg); err != nil {
		tg.HandleError("error sending message: "+err.Error(), "user", user)
		return
	}

	msg = tgbotapi.NewMessage(ANDREW_CHAT_ID, fmt.Sprintf("Зарегистрирован новый пользователь: [%s](%s)", user.Username, "https://t.me/"+user.Username))
	msg.ParseMode = "Markdown"
	if _, err := tg.bot.Request(msg); err != nil {
		tg.HandleError("error sending message: "+err.Error(), "user", user)
		return
	}

}

func (tg *TelegramClient) handleContact(update tgbotapi.Update) {
	chatID := update.FromChat().ID
	user := extractUserInfo(update)
	avatar, _ := tg.getUserProfilePhoto(update)
	user.State = StateWaitingFIO
	user.Avatar = avatar

	if err := tg.store.CreateUser(&user); err != nil {
		tg.HandleError("error while creating user in db: "+err.Error(), "update_id", update.UpdateID)
		return
	}

	msg := tgbotapi.NewMessage(chatID, "Теперь скажи, как к тебе обращаться? Отправь свои фамилию и имя)")
	removeKeyboard := tgbotapi.NewRemoveKeyboard(true)
	msg.ReplyMarkup = removeKeyboard
	if _, err := tg.bot.Request(msg); err != nil {
		tg.HandleError("error sending message: "+err.Error(), "update_id", update.UpdateID)
		return
	}

}

func (tg *TelegramClient) getUserProfilePhoto(update tgbotapi.Update) (string, error) {
	photos, err := tg.bot.GetUserProfilePhotos(tgbotapi.UserProfilePhotosConfig{
		UserID: update.FromChat().ID,
		Limit:  1,
	})
	if err != nil || photos.TotalCount == 0 {
		return "", fmt.Errorf("error getting user profile photos: %v", err)
	}

	fileID := photos.Photos[0][0].FileID
	file, err := tg.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return "", fmt.Errorf("error getting file info: %v", err)
	}
	return fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", os.Getenv("BOT_TOKEN"), file.FilePath), nil

}

func extractUserInfo(update tgbotapi.Update) types.User {
	return types.User{
		Phone:    update.Message.Contact.PhoneNumber,
		Username: update.FromChat().UserName,
		UserID:   update.FromChat().ID,
	}
}

func (tg *TelegramClient) userAccepted(update tgbotapi.Update) {
	callbackQuery := update.CallbackQuery
	chatID, _ := strconv.Atoi(callbackQuery.Data)

	user, err := tg.store.GetUserByID(int64(chatID))
	if err != nil {
		msg := tgbotapi.NewMessage(callbackQuery.From.ID, "Ошибка при подтверждении пользователя")
		if _, err := tg.bot.Request(msg); err != nil {
			tg.HandleError("error sending message: "+err.Error(), "update_id", update.UpdateID)
			return
		}
		tg.HandleError("ошикба при поиске пользователя: "+err.Error(), "update_id", update.UpdateID)
		return
	}

	if err := tg.store.VerifyUser(int64(chatID)); err != nil {
		msg := tgbotapi.NewMessage(callbackQuery.From.ID, "Ошибка при подтверждении пользователя")
		if _, err := tg.bot.Request(msg); err != nil {
			tg.HandleError("error sending message: "+err.Error(), "update_id", update.UpdateID)
			return
		}
		tg.HandleError("ошикба при поиске пользователя: "+err.Error(), "update_id", update.UpdateID)
		return
	}
	callback := tgbotapi.NewCallback(callbackQuery.ID, "Пользователь принят на стажировку")
	if _, err := tg.bot.Request(callback); err != nil {
		tg.HandleError("error sending message: "+err.Error(), "update_id", update.UpdateID)
		return
	}

	del := tgbotapi.NewDeleteMessage(callbackQuery.From.ID, callbackQuery.Message.MessageID)
	if _, err := tg.bot.Send(del); err != nil {
		tg.HandleError("error sending message: "+err.Error(), "update_id", update.UpdateID)
		return
	}

	// Send message to the user that their account has been verified
	msg := tgbotapi.NewMessage(int64(chatID), "Поздравляем, ты принят на обучение🥳 Чтобы преступить к обучению, перейди в наше приложение👇")
	loginButton := tgbotapi.NewInlineKeyboardButtonURL("Начать🔥", "https://t.me/incetro_management_bot/app")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(loginButton))
	msg.ReplyMarkup = keyboard
	if _, err := tg.bot.Request(msg); err != nil {
		tg.HandleError("error sending message: "+err.Error(), "update_id", update.UpdateID)
		return
	}

	if err := tg.store.GiveTasks(int64(chatID), *user.Course); err != nil {
		tg.HandleError("error while giving tasks to user: "+err.Error(), "update_id", update.UpdateID)
	}

}

func (tg *TelegramClient) HandleError(err string, args ...any) {
	if len(args)%2 == 1 {
		args = args[:len(args)-1]
	}
	for i := 0; i < len(args); i += 2 {
		err += fmt.Sprintf("%v=%v", args[i], args[i+1])
	}
	slog.Error(err)

	msg := tgbotapi.NewMessage(MARK_CHAT_ID, err)
	if _, err := tg.bot.Send(msg); err != nil {
		slog.Error("error while handling error: "+err.Error(), "error", err)
	}
}
