package main

import (
	"github.com/ivanov-nikolay/distributed_calculator/internal/app/agent"
)

func main() {
	a := agent.NewApplicationAgent()
	a.RunApplicationAgent()
}
