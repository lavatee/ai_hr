package service

import (
	"github.com/lavatee/ai_hr"
	"github.com/lavatee/ai_hr/internal/repository"
)

type InterviewsService struct {
	repo *repository.Repository
	ai   AIInterviews
}

func NewInterviewsService(repo *repository.Repository, ai AIInterviews) *InterviewsService {
	return &InterviewsService{
		repo: repo,
		ai:   ai,
	}
}

type AIInterviews interface {
	MakeInterview(stack string, difficultyLevel string) (AIInterview, error)
}

func (s *InterviewsService) MakeInterview(stack string, difficultyLevel string, userId int) (int, error) {
	interview, err := s.ai.MakeInterview(stack, difficultyLevel)
	if err != nil {
		return 0, err
	}
	tasks := make([]ai_hr.Task, len(interview.Tasks))
	for index, task := range interview.Tasks {
		var correctAnswerIndex int
		answers := make([]ai_hr.Answer, len(task.Answers))
		for i, answer := range task.Answers {
			answers[i] = ai_hr.Answer{
				Index: i + 1,
				Text:  answer.Text,
			}
			if answer.IsCorrect {
				correctAnswerIndex = i + 1
			}
		}
		tasks[index] = ai_hr.Task{
			Answers:            answers,
			Index:              index + 1,
			Text:               task.Text,
			CorrectAnswerIndex: correctAnswerIndex,
		}
	}
	return s.repo.Interviews.SaveInterview(userId, stack, difficultyLevel, tasks)
}

func (s *InterviewsService) NextInterviewTask(answerId int) (ai_hr.Task, error) {
	return s.repo.Interviews.NextInterviewTask(answerId)
}

func (s *InterviewsService) IsLastTask(answerId int) (bool, error) {
	return s.repo.Interviews.IsLastTask(answerId)
}

func (s *InterviewsService) CompleteInterview(answerId int) ([]ai_hr.Task, error) {
	return s.repo.Interviews.CompleteInterview(answerId)
}

func (s *InterviewsService) GetFirstTask(interviewId int) (ai_hr.Task, error) {
	return s.repo.Interviews.GetFirstTask(interviewId)
}
