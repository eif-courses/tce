package util

import (
	"go.uber.org/zap"
)

var Log *zap.Logger

func InitLogger() {
	logger, err := zap.NewProduction() // or zap.NewDevelopment()
	if err != nil {
		panic("failed to init zap logger: " + err.Error())
	}
	Log = logger
}

func SyncLogger() {
	if Log != nil {
		_ = Log.Sync()
	}
}
