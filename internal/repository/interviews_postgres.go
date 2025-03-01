package repository

type InterviewsPostgres struct {
	db *sqlx.DB
}

func NewInterviewsPostgres(db *sqlx.DB) *InterviewsPostgres {
	return &InterviewsPostgres{
		db: db,
	}
}

func (r *InterviewsPostgres) SaveInterview(userId int, userStack string, difficulty string, tasks []interview.Task) (int, error) {
	var interviewId int
	query := fmt.Sprintf("INSERT INTO %s ()")	
}
