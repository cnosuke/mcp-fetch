package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger initializes the global logger
func InitLogger(debug bool, noLogs bool, logPath string) error {
	var config zap.Config

	if debug {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	} else {
		config = zap.NewProductionConfig()
	}

	// when noLogs is true, only fatal logs are output
	if noLogs {
		config.Level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	}

	if noLogs {
		if logPath != "" {
			// noLogs && len(logPath) > 0: output to logPath only
			config.OutputPaths = []string{logPath}
			config.ErrorOutputPaths = []string{logPath}
		} else {
			// noLogs && len(logPath) == 0: no output
			config.OutputPaths = []string{}
			config.ErrorOutputPaths = []string{}
		}
	} else {
		if logPath != "" {
			// !noLogs && len(logPath) > 0: output to stdout and logPath
			config.OutputPaths = []string{"stdout", logPath}
			config.ErrorOutputPaths = []string{"stderr", logPath}
		} else {
			// !noLogs && len(logPath) == 0: output to stdout and stderr
			config.OutputPaths = []string{"stdout"}
			config.ErrorOutputPaths = []string{"stderr"}
		}
	}

	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return err
	}

	zap.ReplaceGlobals(logger)

	zap.S().Infow("Logger initialized",
		"debug", debug,
		"no_logs", noLogs,
		"log_path", logPath)

	return nil
}

// Sync flushes any buffered log entries
func Sync() error {
	if err := zap.S().Sync(); err != nil {
		return err
	}
	if err := zap.L().Sync(); err != nil {
		return err
	}

	return nil
}
