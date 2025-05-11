package entity

type Expression struct {
	ID     string
	Expr   string
	Status string
	Result *float64
	AST    *ASTNode
	UserID int
}

type ASTNode struct {
	IsLeaf        bool
	Value         float64
	Operator      string
	Left, Right   *ASTNode
	TaskScheduled bool
}
