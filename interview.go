package interview

type Interview struct {
	Id int `db:"id"`
	UserId int `db:"user_id"`
	Stack string `db:"stack"`
	CurrentTaskIndex int `db:"current_task_index"`
}