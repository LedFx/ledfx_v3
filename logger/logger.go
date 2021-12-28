package logger

import (
	"ledfx/config"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

func Init(config config.Config) (logger *zap.SugaredLogger, err error) {
	// First, define our level-handling logic.
	errorLvl := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	standardLvl := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl > zapcore.InfoLevel
	})
	verboseLvl := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		if config.Verbose || config.VeryVerbose {
			return lvl == zapcore.InfoLevel
		}
		return false
	})
	veryVerboseLvl := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		if config.VeryVerbose {
			return lvl <= zapcore.DebugLevel
		}
		return false
	})

	// High-priority output should also go to standard error, and low-priority
	// output should also go to standard out.
	standardSyncer := zapcore.Lock(os.Stdout)
	errorSyncer := zapcore.Lock(os.Stderr)

	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	// Join the outputs, encoders, and level-handling functions into
	// zapcore.Cores, then tee the four cores together.
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, errorSyncer, errorLvl),
		zapcore.NewCore(consoleEncoder, standardSyncer, standardLvl),
		zapcore.NewCore(consoleEncoder, standardSyncer, verboseLvl),
		zapcore.NewCore(consoleEncoder, standardSyncer, veryVerboseLvl),
	)

	// From a zapcore.Core, it's easy to construct a Logger.
	Logger = zap.New(core).Sugar()
	defer func() {
		if err != nil {
			err = Logger.Sync()
		}
	}()

	return
}
