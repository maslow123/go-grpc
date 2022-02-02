package logger

import (
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Log is global logger
	Log *zap.Logger

	// timeFormat is custom Time format
	customTimeFormat string

	// onceInit guarantee initialize logger only once
	onceInit sync.Once
)

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(customTimeFormat))
}

func Init(lvl int, timeFormat string) error {
	var err error
	onceInit.Do(func() {
		// Define level handling logic
		globalLevel := zapcore.Level(lvl)

		highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= globalLevel && lvl < zapcore.ErrorLevel
		})

		lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= globalLevel && lvl < zapcore.ErrorLevel
		})

		consoleInfos := zapcore.Lock(os.Stdout)
		consoleErrors := zapcore.Lock(os.Stderr)

		// Configure console output
		var useCustomTimeFormat bool
		ecfg := zap.NewProductionEncoderConfig()
		if len(timeFormat) > 0 {
			customTimeFormat = timeFormat
			ecfg.EncodeTime = customTimeEncoder
			useCustomTimeFormat = true
		}
		consoleEncoder := zapcore.NewJSONEncoder(ecfg)
		// Join the outputs, encoders, and level handling functions into zapcore
		core := zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
			zapcore.NewCore(consoleEncoder, consoleInfos, lowPriority),
		)

		Log = zap.New(core)
		zap.RedirectStdLog(Log)

		if !useCustomTimeFormat {
			Log.Warn("Time format for logger is not provided - use zap default")
		}
	})

	return err
}
