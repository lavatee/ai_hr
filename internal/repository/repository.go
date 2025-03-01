package repository

type Repository struct {
	Interviews
}

const {
	interviewsTable = "interviews"
	tasksTable = "tasks"
	answersTable = "answers"
}

type Interview interface {
	SaveInterview(userId int, userStack string, difficulty string, tasks []interview.Task) (int, error)
	GetInterviewResults(interviewId int) ([]interview.Task, error)
	NextInterviewQuestion(answerId int) (interview.Question, error)
	DeleteInterview(interviewId int) error
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Interviews: NewInterviewsPostgres(db),
	}
}