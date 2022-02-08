package logs

import log "github.com/sirupsen/logrus"

func ConfigLogLevelToLevel(level int) log.Level {
	switch level {
	case 1:
		return log.InfoLevel
	case 2:
		return log.ErrorLevel
	case 3:
		return log.WarnLevel
	default:
		return log.ErrorLevel
	}
}
