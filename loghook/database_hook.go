package loghook

import (
	cg "github.com/mzpbvsig/common-devices-gateway/internal"
	log "github.com/sirupsen/logrus"
)

type DatabaseHook struct {
	mysqlManager *cg.MysqlManager
}

func NewDatabaseHook(manager *cg.MysqlManager) *DatabaseHook {
	return &DatabaseHook{
		mysqlManager: manager,
	}
}

func (hook *DatabaseHook) Fire(entry *log.Entry) error {
	return hook.mysqlManager.AddLog(entry.Time, entry.Message, entry.Level.String())
}

func (hook *DatabaseHook) Levels() []log.Level {
	return log.AllLevels
}
