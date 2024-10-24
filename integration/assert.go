package integration

import (
	"fmt"
	"log/slog"
	"strings"
)

func NoErr(err error, msg ...string) {
	if err != nil {
		err = fmt.Errorf("assertion failed: %w. %s", err, strings.Join(msg, ", "))

		slog.Error(err.Error())
		panic(err)
	}
}

func Must[T any](val T, err error, msg ...string) T {
	NoErr(err, msg...)
	return val
}
