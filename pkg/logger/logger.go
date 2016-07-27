package logger

import (
	"fmt"
	"io"
	"log"
)

const (
	_ = iota
	LevelError
	LevelWarning
	LevelNotice
	LevelInfo
	LevelDebug
)

var (
	logLevel   = LevelWarning
	levelNames = map[string]int{
		"error":   LevelError,
		"warning": LevelWarning,
		"notice":  LevelNotice,
		"info":    LevelInfo,
		"debug":   LevelDebug,
	}
)

func GetLevelByName(name string) (int, error) {
	level, ok := levelNames[name]
	if !ok {
		return 0, fmt.Errorf("invalid level `%s'", name)
	}

	return level, nil
}

func SetLevel(level int) {
	logLevel = level

	if logLevel < LevelError || logLevel > LevelDebug {
		logLevel = LevelInfo
	}
}

func SetOutput(output io.Writer) {
	log.SetOutput(output)
}

func Log(level int, context, format string, v ...interface{}) {
	var criticity string

	if level > logLevel {
		return
	}

	switch level {
	case LevelError:
		criticity = "ERROR"
	case LevelWarning:
		criticity = "WARNING"
	case LevelNotice:
		criticity = "NOTICE"
	case LevelInfo:
		criticity = "INFO"
	case LevelDebug:
		criticity = "DEBUG"
	}

	log.Printf(
		"%s: %s",
		fmt.Sprintf("%s: %s", criticity, context),
		fmt.Sprintf(format, v...),
	)
}
