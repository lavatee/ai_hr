package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type AIInterviewsService struct {
	ModelName  string
	ApiKey     string
	ApiAddress string
}

func NewAIInterviewsService(modelName string, apiKey string, apiAddress string) *AIInterviewsService {
	return &AIInterviewsService{
		ModelName:  modelName,
		ApiKey:     apiKey,
		ApiAddress: apiAddress,
	}
}

type AIInterviewAnswer struct {
	Text      string `json:"text"`
	IsCorrect bool   `json:"isCorrect"`
}

type AIInterviewQuestion struct {
	Text    string              `json:"text"`
	Answers []AIInterviewAnswer `json:"answers"`
}

type AIInterviewCodeTask struct {
	Text string `json:"text"`
}

type AIInterview struct {
	Questions []AIInterviewQuestion `json:"questions"`
	CodeTask  AIInterviewCodeTask   `json:"codeTask"`
}

type AIContent struct {
	Content string `json:"content"`
}

type AIMessage struct {
	Message AIContent `json:"message"`
}

type AIAnswer struct {
	Choices []AIMessage `json:"choices"`
}

func (s *AIInterviewsService) AIGenerate(message string, callCounter int) (string, error) {
	data := map[string]interface{}{
		"model": s.ModelName,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": message,
			},
		},
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", s.ApiAddress, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	client := &http.Client{}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.ApiKey))
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	logrus.Info(string(body))
	var answer AIAnswer
	if err := json.Unmarshal(body, &answer); err != nil {
		return "", err
	}
	logrus.Info(answer)
	if len(answer.Choices) == 0 {
		if callCounter < 5 {
			return s.AIGenerate(message, callCounter+1)
		} else {
			return "", fmt.Errorf("empty ai answer with query \"%s\"", message)
		}
	} else {
		return answer.Choices[0].Message.Content, nil
	}

}

func (s *AIInterviewsService) MakeInterview(roleName string, mainTechnology string, niceToHave string) (AIInterview, error) {
	message := fmt.Sprintf("Write an interview for the position %s in Russian with the following requirements: \"%s\". The interview must contain 4 questions with 4 answer options: 3 incorrect and 1 correct, as well as 1 coding task (if this position requires coding, otherwise just a question with a full answer (!WITHOUT SOLUTION)), the task should not be too difficult, the user should need no more than 50 lines to solve the task. Try to come up with not very long questions (which should not have very long answers), the maximum length is 255 characters. Send the answer in JSON format, which must contain the questions and codeTask fields. The question object must contain the text and answers fields, the answer object must contain the text and isCorrect fields. The task object must contain the text field.", roleName+" "+mainTechnology, niceToHave)
	var interview AIInterview
	answer, err := s.AIGenerate(message, 1)
	if err != nil {
		return AIInterview{}, err
	}
	if err := json.Unmarshal([]byte(strings.ReplaceAll(strings.ReplaceAll(answer, "```json", ""), "```", "")), &interview); err != nil {
		return AIInterview{}, err
	}

	return interview, nil
}
