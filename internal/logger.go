package internal

import (
	"go.uber.org/zap"
)

var Logger *zap.SugaredLogger

func CreateLogger(debug bool) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	if debug {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	l, err := config.Build()
	if err != nil {
		panic(err)
	}

	Logger = l.Sugar()
}
