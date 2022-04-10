package db

import (
	"context"
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type dlogger struct {
	SlowThreshold         time.Duration
	SourceField           string
	SkipErrRecordNotFound bool
	level                 string
}

func NewLogger(level string) *dlogger {
	skip := false
	if level == "trace" {
		skip = true
	}

	return &dlogger{
		SkipErrRecordNotFound: skip,
		level:                 level,
	}
}

func (l *dlogger) LogMode(gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *dlogger) Info(ctx context.Context, s string, args ...interface{}) {
	log.WithContext(ctx).Debugf(s, args)
}

func (l *dlogger) Warn(ctx context.Context, s string, args ...interface{}) {
	log.WithContext(ctx).Warnf(s, args)
}

func (l *dlogger) Error(ctx context.Context, s string, args ...interface{}) {
	log.WithContext(ctx).Errorf(s, args)
}

func (l *dlogger) Debug(ctx context.Context, s string, args ...interface{}) {
	log.WithContext(ctx).Debugf(s, args)
}

func (l *dlogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, _ := fc()
	fields := log.Fields{}
	if l.SourceField != "" {
		fields[l.SourceField] = utils.FileWithLineNum()
	}
	if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) && l.SkipErrRecordNotFound) {
		fields[log.ErrorKey] = err
		log.WithContext(ctx).WithFields(fields).Errorf("%s [%s]", sql, elapsed)
		return
	}

	if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		log.WithContext(ctx).WithFields(fields).Warnf("%s [%s]", sql, elapsed)
		return
	}

	log.WithContext(ctx).WithFields(fields).Tracef("%s [%s]", sql, elapsed)
}
