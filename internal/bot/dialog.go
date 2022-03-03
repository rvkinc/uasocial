package bot

import (
	"fmt"
	tg "github.com/go-telegram-bot-api/telegram-bot-api"
	"go.uber.org/zap"
	"strings"
)

type MessageHandler struct {
	Api        *tg.BotAPI
	L          *zap.Logger
	Translator Translator

	state map[int64]*dialog
}

func NewMessageHandler(api *tg.BotAPI, l *zap.Logger) *MessageHandler {
	tr, err := NewTranslator()
	if err != nil {
		panic(err)
	}

	m := &MessageHandler{
		Api:        api,
		L:          l,
		Translator: tr,

		state: make(map[int64]*dialog),
	}

	return m
}

// bot commands
const (
	startCmd = "start"
)

type role int

const (
	roleVolunteer role = iota + 1
	roleSeeker
)

type category int

const (
	categoryFood category = iota
	categoryMeds
	categoryClothes
	categoryApartments
	categoryTransport
	categoryOther
)

type handler func(*Update) error

type volunteer struct {
	categories       []category
	categoryKeyboard []checkboxKeyboard
}

type seeker struct {
}

type dialog struct {
	role role
	next handler

	// either one is populated during the dialog
	volunteer *volunteer
	seeker    *seeker
}

func (m *MessageHandler) Handle(_ *tg.BotAPI, u *Update) {

	if u.Message != nil && u.Message.IsCommand() {
		switch u.Message.Command() {
		case startCmd:
			err := m.userRoleRequest(u)
			if err != nil {
				m.L.Error("handle start cmd", zap.Error(err))
			}
			return
		default:
			_, _ = m.Api.Send(tg.NewMessage(u.chatID(), "Error"))
			return
		}
	}

	dialog, ok := m.state[u.chatID()]
	if !ok {
		err := m.userRoleRequest(u)
		if err != nil {
			m.L.Error("handle user role request", zap.Error(err))
		}
		return
	}

	err := dialog.next(u)
	if err != nil {
		m.L.Error("handle request", zap.Error(err))
	}
}

func (m *MessageHandler) userRoleRequest(u *Update) error {

	msg := tg.NewMessage(u.chatID(), m.Translator.Translate(userRoleRequestTranslation, UALang))
	msg.ReplyMarkup = tg.ReplyKeyboardMarkup{
		OneTimeKeyboard: false,
		Keyboard: [][]tg.KeyboardButton{
			{tg.KeyboardButton{Text: m.Translator.Translate(btnOptionUserRoleSeeker, UALang)}},
			{tg.KeyboardButton{Text: m.Translator.Translate(btnOptionUserRoleVolunteer, UALang)}},
		},
	}

	_, err := m.Api.Send(msg)
	if err != nil {
		return err
	}

	m.state[u.chatID()] = &dialog{next: m.handleUserRoleReply}
	return nil
}

func (m *MessageHandler) handleUserRoleReply(u *Update) error {
	switch u.Message.Text {
	case m.Translator.Translate(btnOptionUserRoleSeeker, UALang):
	case m.Translator.Translate(btnOptionUserRoleVolunteer, UALang):
		d := m.state[u.chatID()]
		d.role = roleVolunteer
		d.volunteer = new(volunteer)
		d.volunteer.categoryKeyboard = []checkboxKeyboard{
			{text: m.Translator.Translate(categoryFoodTr, UALang), category: categoryFood, checked: false},
			{text: m.Translator.Translate(categoryMedsTr, UALang), category: categoryMeds, checked: false},
			{text: m.Translator.Translate(categoryClothesTr, UALang), category: categoryClothes, checked: false},
			{text: m.Translator.Translate(categoryApartmentsTr, UALang), category: categoryApartments, checked: false},
			{text: m.Translator.Translate(categoryThransportTr, UALang), category: categoryTransport, checked: false},
			{text: m.Translator.Translate(categoryOtherTr, UALang), category: categoryOther, checked: false},
		}
		// d.volunteer.categoryKeyboard = [][]tg.KeyboardButton{
		// 	{tg.KeyboardButton{Text: uncheckedCheckbox + " " + m.Translator.Translate(categoryFoodTr, UALang)}},
		// 	{tg.KeyboardButton{Text: uncheckedCheckbox + " " + m.Translator.Translate(categoryMedsTr, UALang)}},
		// 	{tg.KeyboardButton{Text: uncheckedCheckbox + " " + m.Translator.Translate(categoryClothesTr, UALang)}},
		// 	{tg.KeyboardButton{Text: uncheckedCheckbox + " " + m.Translator.Translate(categoryApartmentsTr, UALang)}},
		// 	{tg.KeyboardButton{Text: uncheckedCheckbox + " " + m.Translator.Translate(categoryThransportTr, UALang)}},
		// 	{tg.KeyboardButton{Text: uncheckedCheckbox + " " + m.Translator.Translate(categoryOtherTr, UALang)}},
		// }

		msg := tg.NewMessage(u.chatID(), m.Translator.Translate(userRoleRequestTranslation, UALang))
		msg.ReplyMarkup = tg.ReplyKeyboardMarkup{
			OneTimeKeyboard: false,
			Keyboard:        checkboxKeyboardLayout(d.volunteer.categoryKeyboard),
		}

		_, err := m.Api.Send(msg)
		if err != nil {
			return err
		}

		m.state[u.chatID()].next = m.handleVolunteerCategoryCheckbox
	default:
		_, err := m.Api.Send(tg.NewMessage(u.chatID(), m.Translator.Translate(errorChooseOption, UALang)))
		if err != nil {
			return err
		}
	}

	return nil
}

const (
	uncheckedCheckbox = "🔲"
	checkedCheckbox   = "✅"
)

type checkboxKeyboard struct {
	text     string
	category category
	checked  bool
}

func (k *checkboxKeyboard) keyboardButton() tg.KeyboardButton {
	var checkbox = uncheckedCheckbox
	if k.checked {
		checkbox = checkedCheckbox
	}
	return tg.KeyboardButton{Text: checkbox + " " + k.text}
}

func (k *checkboxKeyboard) invert(text string) bool {
	if strings.Contains(text, k.text) {
		k.checked = !k.checked
		return true
	}

	return false
}

func checkboxKeyboardLayout(k []checkboxKeyboard) [][]tg.KeyboardButton {
	layout := make([][]tg.KeyboardButton, 0, len(k))
	for _, key := range k {
		layout = append(layout, []tg.KeyboardButton{key.keyboardButton()})
	}
	return layout
}

func (m *MessageHandler) handleVolunteerCategoryCheckbox(u *Update) error {
	d := m.state[u.chatID()]
	var checked category
	for _, keyboard := range d.volunteer.categoryKeyboard {
		if keyboard.invert(u.Message.Text) {
			d.volunteer.categories = append(d.volunteer.categories, keyboard.category)
			checked = keyboard.category
			break
		}
	}

	msg := tg.NewMessage(u.chatID(), fmt.Sprintf("%d", checked))
	msg.ReplyMarkup = tg.ReplyKeyboardMarkup{
		OneTimeKeyboard: false,
		Keyboard:        checkboxKeyboardLayout(d.volunteer.categoryKeyboard),
	}

	_, err := m.Api.Send(msg)
	return err
}

// func (m *MessageHandler) contactPhoneRequest(u *Update) {
// 	msg := tg.NewMessage(u.chatID(), "Your contact number")
// 	msg.ReplyMarkup = tg.ReplyKeyboardMarkup{
// 		OneTimeKeyboard: true,
// 		Keyboard: [][]tg.KeyboardButton{
// 			{
// 				tg.KeyboardButton{
// 					Text:           "Contact number",
// 					RequestContact: true,
// 				},
// 			},
// 		},
// 	}
//
// 	_, err := m.Api.Send(msg)
// 	if err != nil {
// 		m.L.Error("send message", zap.Error(err))
// 	}
// }
