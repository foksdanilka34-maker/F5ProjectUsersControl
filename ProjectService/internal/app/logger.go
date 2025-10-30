package app

import (
	"log"
	"os"
)

var Logger *log.Logger

func init() {
	Logger = log.New(os.Stdout, "project-service: ", log.LstdFlags)
}
