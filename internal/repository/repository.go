package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/lavatee/ai_hr"
)

type Repository struct {
	Interviews
}

const (
	interviewsTable = "interviews"
	tasksTable      = "tasks"
	answersTable    = "answers"
)

type Interviews interface {
	SaveInterview(userId int, userStack string, difficulty string, tasks []ai_hr.Task) (int, error)
	CompleteInterview(answerId int) ([]ai_hr.Task, error)
	NextInterviewTask(answerId int) (ai_hr.Task, error)
	IsLastTask(answerId int) (bool, error)
	GetFirstTask(interviewId int) (ai_hr.Task, error)
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Interviews: NewInterviewsPostgres(db),
	}
}
