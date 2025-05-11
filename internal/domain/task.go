package domain

type Task struct {
	ID            string  `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
}

type TaskResult struct {
	ID     string  `json:"id"`
	Result float64 `json:"result"`
}

type TaskRepository interface {
	FetchTask() (*Task, error)
	SubmitResult(result TaskResult) error
	ReportInvalidTask(taskID string, operation string) error
}
