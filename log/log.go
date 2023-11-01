package log

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"os"
	"time"
)

var (
	timestampFormat = log.TimestampFormat(
		func() time.Time { return time.Now().UTC() },
		"2006-01-02T15:04:05.000Z07:00",
	)
)

type Config struct {
	Level  *AllowedLevel
	Format *AllowedFormat
}

type AllowedLevel struct {
	s string
	o level.Option
}

type AllowedFormat struct {
	s string
}

func New(config *Config) log.Logger {
	var l log.Logger
	if config.Format != nil && config.Format.s == "json" {
		l = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	} else {
		l = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	}

	if config.Level != nil {
		l = log.With(l, "ts", timestampFormat, "caller", log.Caller(5))
		l = level.NewFilter(l, config.Level.o)
	} else {
		l = log.With(l, "ts", timestampFormat, "caller", log.DefaultCaller)
	}
	return l
}
