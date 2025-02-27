package interview

type Answer struct {
	Id int `db:"id"`
	TaskId int `db:"task_id"`
	Index int `db:"index"`
	Text string `db:"text"`
}