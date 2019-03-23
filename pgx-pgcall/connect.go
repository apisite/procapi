package pgxpgcall

import (
	"github.com/jackc/pgx" // gopkg failed because "internal" lib used
	"github.com/pkg/errors"
	"gopkg.in/birkirb/loggers.v1"
)

func initPool(cfg Config, log loggers.Contextual) (*pgx.ConnPoolConfig, error) {
	dbConf, err := pgx.ParseEnvLibpq()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse environment")
	}
	level, err := pgx.LogLevelFromString(cfg.LogLevel)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to parse log level "+cfg.LogLevel)
	}
	dbConf.LogLevel = int(level)
	dbConf.Logger = Logger{l: log}

	config := pgx.ConnPoolConfig{
		ConnConfig:     dbConf,
		MaxConnections: cfg.Workers,
		AfterConnect: func(conn *pgx.Conn) error {
			// This code does not included in coverage report but it is called (and tested) in db_test.do
			if cfg.Schema != "" {
				log.Debugf("DB searchpath: (%s)", cfg.Schema)
				_, err = conn.Exec("set search_path = " + cfg.Schema)
			}
			log.Debugf("Added DB connection")
			return err
		},
	}
	return &config, nil
}

// Logger holds database logging via given logger
type Logger struct {
	l loggers.Contextual
}

// Log message via logger
func (l Logger) Log(level pgx.LogLevel, msg string, data map[string]interface{}) {

	switch level {
	case pgx.LogLevelTrace:
		data["PGX_LOG_LEVEL"] = level
		l.l.WithFields(data).Debug(msg)
	case pgx.LogLevelDebug:
		l.l.WithFields(data).Debug(msg)
	case pgx.LogLevelInfo:
		l.l.WithFields(data).Info(msg)
	case pgx.LogLevelWarn:
		l.l.WithFields(data).Warn(msg)
	case pgx.LogLevelError:
		l.l.WithFields(data).Error(msg)
	default:
		data["INVALID_PGX_LOG_LEVEL"] = level
		l.l.WithFields(data).Error(msg)
	}
}
