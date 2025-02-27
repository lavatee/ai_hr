package interview

type Task struct {
	Id int `db:"id"`
	InterviewId int `db:"interview_id"`
	Index int `db:"index"`
	Text string `db:"text"`
	CorrectAnswerIndex int `db:"correct_answer_index"`
	IsCorrect bool `db:"isCorrect"`
	Answers []Answer
}