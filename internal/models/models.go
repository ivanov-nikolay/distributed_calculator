package models

// Expression описание математического выражения
type Expression struct {
	// ID выражения
	ID string `json:"id"`
	// Expr математическое выражение
	Expr string `json:"expression"`
	// Status текущее состояние вычисления математического выражения
	Status string `json:"status"`
	// Result результат вычисления математического выражения
	Result float64 `json:"result"`
}

// Task описание задачи для агента
type Task struct {
	// ID задачи
	ID string `json:"id"`
	// Arg1 первый аргумент
	Arg1 float64 `json:"arg1"`
	// Arg2 второй аргумент
	Arg2 float64 `json:"arg2"`
	// Operator математическое действие ("+", "-", "*", "/")
	Operation string `json:"operation"`
	// OperationTime время выполнения операции в миллисекундах
	OperationTime int `json:"operation_time"`
}

// TaskResult описание результата выполнения задачи
type TaskResult struct {
	// ID задачи
	ID string `json:"id"`
	// Result результат выполнения задачи
	Result float64 `json:"result"`
}

// TaskReceived принятая задача агентом
type TaskReceived struct {
	Task Task `json:"task"`
}
