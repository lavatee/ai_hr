package endpoint

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/sirupsen/logrus"
)

var userStates = map[int64]string{}

var messagesToDelete = map[int64][]int{}

func (e *Endpoint) HandleCallback(bot *telego.Bot, query telego.CallbackQuery) {
	callbacksMap := map[string]func(bot *telego.Bot, query telego.CallbackQuery){
		"makeInterview": e.StartMakingInterview,
		"generate":      e.GenerateInterview,
		"answer":        e.NextInterviewTask,
	}
	logrus.Infof("callback query: %s", query.Data)
	callbacksMap[strings.Split(query.Data, "_")[0]](bot, query)
}

func (e *Endpoint) StartMakingInterview(bot *telego.Bot, query telego.CallbackQuery) {
	userStates[query.From.ID] = "making"
	_, _ = bot.SendMessage(tu.Message(tu.ID(query.From.ID), "Напишите свое направление (например Backend Python, Backend, или просто Python). На основе вашего направления бот сгенерирует собеседование"))
	for _, messageId := range messagesToDelete[query.From.ID] {
		_ = bot.DeleteMessage(tu.Delete(tu.ID(query.From.ID), messageId))
	}
}

func (e *Endpoint) HandleAnyMessage(bot *telego.Bot, update telego.Update) {
	if _, ok := userStates[update.Message.From.ID]; ok {
		statesMap := map[string]func(bot *telego.Bot, update telego.Update){
			"making": e.HandleMakingState,
		}
		statesMap[userStates[update.Message.From.ID]](bot, update)
	} else {
		_, _ = bot.SendMessage(tu.Message(tu.ID(update.Message.From.ID), "Неизвестное сообщение"))
	}
}

func (e *Endpoint) HandleMakingState(bot *telego.Bot, update telego.Update) {
	makingKeyboard := tu.InlineKeyboard(tu.InlineKeyboardRow(tu.InlineKeyboardButton("Junior").WithCallbackData(fmt.Sprintf("generate_%s_with easy tasks", strings.ReplaceAll(update.Message.Text, "_", " ")))), tu.InlineKeyboardRow(tu.InlineKeyboardButton("Middle").WithCallbackData(fmt.Sprintf("generate_%s_with hard tasks", strings.ReplaceAll(update.Message.Text, "_", " ")))), tu.InlineKeyboardRow(tu.InlineKeyboardButton("Senior").WithCallbackData(fmt.Sprintf("generate_%s_WITH VERY HARD TASKS", strings.ReplaceAll(update.Message.Text, "_", " ")))))
	message, _ := bot.SendMessage(tu.Message(tu.ID(update.Message.Chat.ID), "Выбери сложность собеседования").WithReplyMarkup(makingKeyboard))
	userMessagesToDelete := messagesToDelete[update.Message.From.ID]
	userMessagesToDelete = append(userMessagesToDelete, message.MessageID)
	messagesToDelete[update.Message.From.ID] = userMessagesToDelete
}

func (e *Endpoint) GenerateInterview(bot *telego.Bot, query telego.CallbackQuery) {
	for _, messageId := range messagesToDelete[query.From.ID] {
		_ = bot.DeleteMessage(tu.Delete(tu.ID(query.From.ID), messageId))
	}
	_, _ = bot.SendMessage(tu.Message(tu.ID(query.From.ID), "Генерация собеседования началась!"))
	stack := strings.Split(query.Data, "_")[1]
	difficulty := strings.Split(query.Data, "_")[2]
	interviewId, err := e.services.Interviews.MakeInterview(stack, difficulty, int(query.From.ID))
	if err != nil {
		_, _ = bot.SendMessage(tu.Message(tu.ID(query.From.ID), "Произошла ошибка при генерации собеседования"))
		logrus.Errorf("interview generating error: %s", err.Error())
		return
	}
	firstTask, err := e.services.Interviews.GetFirstTask(interviewId)
	if err != nil {
		_, _ = bot.SendMessage(tu.Message(tu.ID(query.From.ID), "Произошла ошибка при генерации собеседования"))
		logrus.Errorf("first task getting error: %s", err.Error())
		return
	}
	taskAnswers := make([][]telego.InlineKeyboardButton, len(firstTask.Answers))
	for i, answer := range firstTask.Answers {
		taskAnswers[i] = tu.InlineKeyboardRow(tu.InlineKeyboardButton(answer.Text).WithCallbackData(fmt.Sprintf("answer_%d", answer.Id)))
	}
	answersKeyboard := tu.InlineKeyboard(taskAnswers...)
	message, err := bot.SendMessage(tu.Message(tu.ID(query.From.ID), firstTask.Text).WithReplyMarkup(answersKeyboard))
	userMessagesToDelete := messagesToDelete[query.From.ID]
	userMessagesToDelete = append(userMessagesToDelete, message.MessageID)
	messagesToDelete[query.From.ID] = userMessagesToDelete
}

func (e *Endpoint) NextInterviewTask(bot *telego.Bot, query telego.CallbackQuery) {

	answerId, err := strconv.Atoi(strings.Split(query.Data, "_")[1])
	if err != nil {
		_, _ = bot.SendMessage(tu.Message(tu.ID(query.From.ID), "Произошла ошибка при отправке ответа"))
		logrus.Errorf("answer to int error: %s", err.Error())
		return
	}
	isLastTask, err := e.services.Interviews.IsLastTask(answerId)
	if err != nil {
		_, _ = bot.SendMessage(tu.Message(tu.ID(query.From.ID), "Произошла ошибка при проверки ответа"))
		logrus.Errorf("is last task error: %s", err.Error())
		return
	}
	if isLastTask {
		results, err := e.services.Interviews.CompleteInterview(answerId)
		if err != nil {
			_, _ = bot.SendMessage(tu.Message(tu.ID(query.From.ID), "Произошла ошибка при завершения собеседования"))
			logrus.Errorf("interview results error: %s", err.Error())
			return
		}
		resultsMessage := "Результаты: \n"
		for _, result := range results {
			emojiMap := map[bool]string{
				true:  "✅",
				false: "🚫",
			}
			resultsMessage += fmt.Sprintf("%s: %s\n", result.Text, emojiMap[result.IsCorrect])
		}
		_, _ = bot.SendMessage(tu.Message(tu.ID(query.From.ID), resultsMessage))
		for _, messageId := range messagesToDelete[query.From.ID] {
			_ = bot.DeleteMessage(tu.Delete(tu.ID(query.From.ID), messageId))
		}
		return
	}
	task, err := e.services.Interviews.NextInterviewTask(answerId)
	if err != nil {
		_, _ = bot.SendMessage(tu.Message(tu.ID(query.From.ID), "Произошла ошибка при проверки ответа"))
		logrus.Errorf("answer task error: %s", err.Error())
		return
	}
	for _, messageId := range messagesToDelete[query.From.ID] {
		_ = bot.DeleteMessage(tu.Delete(tu.ID(query.From.ID), messageId))
	}
	taskAnswers := make([][]telego.InlineKeyboardButton, len(task.Answers))
	for i, answer := range task.Answers {
		taskAnswers[i] = tu.InlineKeyboardRow(tu.InlineKeyboardButton(answer.Text).WithCallbackData(fmt.Sprintf("answer_%d", answer.Id)))
	}
	answersKeyboard := tu.InlineKeyboard(taskAnswers...)
	message, err := bot.SendMessage(tu.Message(tu.ID(query.From.ID), task.Text).WithReplyMarkup(answersKeyboard))
	userMessagesToDelete := messagesToDelete[query.From.ID]
	userMessagesToDelete = append(userMessagesToDelete, message.MessageID)
	messagesToDelete[query.From.ID] = userMessagesToDelete
}

func (e *Endpoint) HandleStart(bot *telego.Bot, update telego.Update) {
	startKeyboard := tu.InlineKeyboard(tu.InlineKeyboardRow(tu.InlineKeyboardButton("Составить собеседование").WithCallbackData("makeInterview")))
	message, _ := bot.SendMessage(tu.Message(tu.ID(update.Message.Chat.ID), "Привет! Это бот-тренер по IT собеседованиям. Нажми на кнопку \"Составить собеседование\"").WithReplyMarkup(startKeyboard))
	userMessagesToDelete := messagesToDelete[update.Message.From.ID]
	userMessagesToDelete = append(userMessagesToDelete, message.MessageID)
	messagesToDelete[update.Message.From.ID] = userMessagesToDelete
}
