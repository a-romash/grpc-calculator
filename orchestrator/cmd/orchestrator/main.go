package main

import (
	"calculator/internal/orchestrator/server"
	"calculator/pkg/config"
)

func main() {
	config.Init()
	newOrchestrator := server.New()
	newOrchestrator.Run()
}
