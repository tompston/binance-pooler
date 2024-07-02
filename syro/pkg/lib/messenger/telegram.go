// Package messenger holds a telegram bot that can be used to send messages
package messenger

import (
	"fmt"
	"strconv"
	"syro/pkg/app/settings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type bot struct {
	token    string
	chatId   int64
	instance *tgbotapi.BotAPI
}

func (b *bot) Instance() *tgbotapi.BotAPI { return b.instance }

func NewTelegramBot(token string, chatId int) (*bot, error) {
	if chatId == 0 {
		return nil, fmt.Errorf("telegram chat id is not specified")
	}

	instance, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	if instance == nil {
		return nil, fmt.Errorf("failed to initialize telegram bot")
	}

	return &bot{token, int64(chatId), instance}, nil
}

// NewEnvTelegramBot creates a new telegram bot instance based on the environment variables.
func NewEnvTelegramBot(envTokenKey, envChatIdKey string) (*bot, error) {
	token := settings.GetEnvVar(envTokenKey)
	if token == "" {
		return nil, fmt.Errorf("telegram token is empty")
	}

	chatId, err := strconv.Atoi(settings.GetEnvVar(envChatIdKey))
	if err != nil {
		return nil, err
	}

	if chatId == 0 {
		return nil, fmt.Errorf("telegram chat id is not specified")
	}

	return NewTelegramBot(token, chatId)
}

// Send a message to the telegram chat.
func (b *bot) Send(msg string) error {
	if b.instance == nil {
		return fmt.Errorf("telegram bot is not initialized")
	}

	loc, err := time.LoadLocation("Europe/Riga")
	if err != nil {
		return fmt.Errorf("failed to load Europe/Riga location: %v", err)
	}

	rigaTime := time.Now().In(loc).Format("2006-01-02 15:04")
	content := fmt.Sprintf("%s ~ %v", rigaTime, msg)

	message := tgbotapi.NewMessage(b.chatId, content)
	_, err = b.instance.Send(message)
	return err
}

func (b *bot) SendAttachment(attachment []byte, name string) error {
	if b.instance == nil {
		return fmt.Errorf("telegram bot is not initialized")
	}

	if len(attachment) == 0 || attachment == nil {
		return fmt.Errorf("attachment is empty")
	}

	message := tgbotapi.NewDocumentUpload(b.chatId, tgbotapi.FileBytes{
		Name:  name,
		Bytes: attachment,
	})

	_, err := b.instance.Send(message)
	return err
}

type TelegramAttachment struct {
	Attachment []byte
	Name       string
}

// This is basically just a shorthand to call the Send method,
// so that the caller doesn't have to check for errors when
// initiaizing the bot + avoid panics if the bot is nil.
func SendTgMessage(envTokenKey, envChatIdKey, msg string, attachment ...TelegramAttachment) error {
	bot, err := NewEnvTelegramBot(envTokenKey, envChatIdKey)
	if err != nil {
		return err
	}

	if err := bot.Send(msg); err != nil {
		return err
	}

	if len(attachment) == 1 && len(attachment[0].Attachment) != 0 {
		at := attachment[0]
		return bot.SendAttachment(at.Attachment, at.Name)
	}

	return nil
}
