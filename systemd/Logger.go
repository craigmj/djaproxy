package systemd

import (
	logger `log`
)

type Logger struct {}

func (l *Logger) Error(err error) {
	if nil==err {
		return
	}
	logger.Println(err.Error())
}

var log *Logger


