package service

import (
	"github.com/lavatee/ai_hr"
	"github.com/lavatee/ai_hr/internal/repository"
)

type Service struct {
	Interviews
}

type Interviews interface {
	MakeInterview(stack string, difficultyLevel string, userId int) (int, error)
	NextInterviewTask(answerId int) (ai_hr.Task, error)
	CompleteInterview(answerId int) ([]ai_hr.Task, error)
	GetFirstTask(interviewId int) (ai_hr.Task, error)
	IsLastTask(answerId int) (bool, error)
}

func NewService(repo *repository.Repository, aiModel string, apiAddress string, apiKey string) *Service {
	return &Service{
		Interviews: NewInterviewsService(repo, NewAIInterviewsService(aiModel, apiKey, apiAddress)),
	}
}
