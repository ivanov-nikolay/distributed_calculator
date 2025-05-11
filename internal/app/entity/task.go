package entity

type Task struct {
	ID            string
	ExprID        string
	Arg1          float64
	Arg2          float64
	Operation     string
	OperationTime int
	Node          *ASTNode
}
