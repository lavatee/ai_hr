package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lavatee/ai_hr"
	"github.com/sirupsen/logrus"
)

type InterviewsPostgres struct {
	db *sqlx.DB
}

func NewInterviewsPostgres(db *sqlx.DB) *InterviewsPostgres {
	return &InterviewsPostgres{
		db: db,
	}
}

const lastTaskIndex = 5

func (r *InterviewsPostgres) GetFirstTask(interviewId int) (ai_hr.Task, error) {
	return r.GetNextTask(AnswerInfo{InterviewId: interviewId, TaskIndex: 0})
}

func (r *InterviewsPostgres) SaveInterview(userId int, userStack string, difficulty string, tasks []ai_hr.Task) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	var interviewId int
	query := fmt.Sprintf("INSERT INTO %s (user_id, stack, difficulty, current_task_index) VALUES ($1, $2, $3, $4) RETURNING id", interviewsTable)
	row := tx.QueryRow(query, 1, userStack, difficulty, 1)
	if err := row.Scan(&interviewId); err != nil {
		tx.Rollback()
		return 0, err
	}
	for _, task := range tasks {
		var taskId int
		query := fmt.Sprintf("INSERT INTO %s (interview_id, text, is_correct, correct_answer_index, index) VALUES ($1, $2, $3, $4, $5) RETURNING id", tasksTable)
		row := tx.QueryRow(query, interviewId, task.Text, false, task.CorrectAnswerIndex, task.Index)
		if err := row.Scan(&taskId); err != nil {
			tx.Rollback()
			return 0, err
		}
		answersArgs := make([]interface{}, 0)
		argsCounter := 0
		answersQuery := fmt.Sprintf("INSERT INTO %s (task_id, text, index) VALUES ", answersTable)
		for i, answer := range task.Answers {
			answersQuery += fmt.Sprintf("($%d, $%d, $%d)", argsCounter+1, argsCounter+2, argsCounter+3)
			argsCounter += 3
			if i < len(task.Answers)-1 {
				answersQuery += ", "
			}
			answersArgs = append(answersArgs, taskId, answer.Text, answer.Index)
		}
		if len(task.Answers) != 0 {
			_, err := tx.Exec(answersQuery, answersArgs...)
			if err != nil {
				tx.Rollback()
				return 0, err
			}
		}

	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, err
	}
	return interviewId, nil

}

type AnswerInfo struct {
	AnswerIndex        int `db:"index"`
	TaskId             int `db:"id"`
	CorrectAnswerIndex int `db:"correct_answer_index"`
	InterviewId        int `db:"interview_id"`
	TaskIndex          int `db:"task_index"`
}

func (r *InterviewsPostgres) GetAnswerInfo(answerId int) (AnswerInfo, error) {
	var answerInfo AnswerInfo
	query := fmt.Sprintf("SELECT a.index, t.id, t.correct_answer_index, t.interview_id, t.index AS task_index FROM %s a JOIN %s t ON a.task_id = t.id WHERE a.id = $1", answersTable, tasksTable)
	if err := r.db.Get(&answerInfo, query, answerId); err != nil {
		return answerInfo, err
	}
	return answerInfo, nil
}

func (r *InterviewsPostgres) GetNextTask(answerInfo AnswerInfo) (ai_hr.Task, error) {
	var task ai_hr.Task
	query := fmt.Sprintf("SELECT id, index, text FROM %s WHERE interview_id = $1 AND index = $2", tasksTable)
	if err := r.db.Get(&task, query, answerInfo.InterviewId, answerInfo.TaskIndex+1); err != nil {
		return ai_hr.Task{}, err
	}
	var taskAnswers []ai_hr.Answer
	query = fmt.Sprintf("SELECT id, index, text FROM %s WHERE task_id = $1", answersTable)
	if err := r.db.Select(&taskAnswers, query, task.Id); err != nil {
		return ai_hr.Task{}, err
	}
	task.Answers = taskAnswers
	return task, nil
}

func (r *InterviewsPostgres) AnswerTask(answerInfo AnswerInfo) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	query := fmt.Sprintf("UPDATE %s SET is_correct = $1 WHERE id = $2", tasksTable)
	logrus.Infof("correct answer index: %d, user answer: %d", answerInfo.CorrectAnswerIndex, answerInfo.AnswerIndex)
	_, err = tx.Exec(query, answerInfo.AnswerIndex == answerInfo.CorrectAnswerIndex, answerInfo.TaskId)
	if err != nil {
		tx.Rollback()
		return err
	}
	query = fmt.Sprintf("UPDATE %s SET current_task_index = current_task_index + 1 WHERE id = $1", interviewsTable)
	_, err = tx.Exec(query, answerInfo.InterviewId)
	if err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func (r *InterviewsPostgres) NextInterviewTask(answerId int) (ai_hr.Task, error) {
	answerInfo, err := r.GetAnswerInfo(answerId)
	if err != nil {
		return ai_hr.Task{}, err
	}
	if err := r.AnswerTask(answerInfo); err != nil {
		return ai_hr.Task{}, err
	}
	return r.GetNextTask(answerInfo)
}

func (r *InterviewsPostgres) IsLastTask(answerId int) (bool, error) {
	var taskIndex int
	query := fmt.Sprintf("SELECT t.index FROM %s a JOIN %s t ON a.task_id = t.id WHERE a.id = $1", answersTable, tasksTable)
	row := r.db.QueryRow(query, answerId)
	if err := row.Scan(&taskIndex); err != nil {
		return false, err
	}
	return taskIndex == lastTaskIndex, nil
}

func (r *InterviewsPostgres) CompleteInterview(answerId int) ([]ai_hr.Task, error) {
	answerInfo, err := r.GetAnswerInfo(answerId)
	if err != nil {
		return []ai_hr.Task{}, err
	}
	if err := r.AnswerTask(answerInfo); err != nil {
		return []ai_hr.Task{}, err
	}
	var results []ai_hr.Task
	query := fmt.Sprintf("SELECT text, is_correct FROM %s WHERE interview_id = $1", tasksTable)
	if err := r.db.Select(&results, query, answerInfo.InterviewId); err != nil {
		return []ai_hr.Task{}, err
	}
	if err := r.DeleteInterview(answerInfo.InterviewId); err != nil {
		return []ai_hr.Task{}, err
	}
	return results, nil
}

func (r *InterviewsPostgres) DeleteInterview(interviewId int) error {
	var tasks []ai_hr.Task
	query := fmt.Sprintf("SELECT id FROM %s WHERE interview_id = $1", tasksTable)
	if err := r.db.Select(&tasks, query, interviewId); err != nil {
		return err
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	query = fmt.Sprintf("DELETE FROM %s WHERE", answersTable)
	argsCounter := 0
	for i := range tasks {
		query += fmt.Sprintf(" task_id = $%d", argsCounter+1)
		argsCounter++
		if i != len(tasks)-1 {
			query += " OR"
		}
	}
	logrus.Info(query)
	args := make([]interface{}, len(tasks))
	for i, task := range tasks {
		args[i] = task.Id
	}
	_, err = tx.Exec(query, args...)
	if err != nil {
		tx.Rollback()
		return err
	}
	query = fmt.Sprintf("DELETE FROM %s WHERE interview_id = $1", tasksTable)
	_, err = tx.Exec(query, interviewId)
	if err != nil {
		tx.Rollback()
		return err
	}
	query = fmt.Sprintf("DELETE FROM %s WHERE id = $1", interviewsTable)
	_, err = tx.Exec(query, interviewId)
	if err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
