package app

import (
	"github.com/foksdanilka34-maker/F5ProjectUsersControl/pkg/logger"
	"go.uber.org/zap"
)

var Logger *zap.Logger

func init() {
	var err error
	Logger, err = logger.New("project-service")
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
}
