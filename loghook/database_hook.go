package loghook

import (
	"fmt"

	"github.com/mzpbvsig/common-devices-gateway/internal"
	log "github.com/sirupsen/logrus"
)

type DatabaseHook struct {
	mysqlManager *internal.MysqlManager
}

func NewDatabaseHook(manager *internal.MysqlManager) *DatabaseHook {
	return &DatabaseHook{
		mysqlManager: manager,
	}
}

func (hook *DatabaseHook) Fire(entry *log.Entry) error {
	if hook.mysqlManager != nil {
		return hook.mysqlManager.AddLog(entry.Time, entry.Message, entry.Level.String())
	}
	return fmt.Errorf("mysql manager is nil pointer")
}

func (hook *DatabaseHook) Levels() []log.Level {
	return log.AllLevels
}
