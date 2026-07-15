package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger
var Sugar *zap.SugaredLogger

// Initialize инициализирует синглтон логера с необходимым уровнем логирования.
func Initialize(level string) error {
	Log = zap.Must(zap.NewProduction())

	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	encoderCfg := zapcore.EncoderConfig{
		MessageKey:       "msg",
		LevelKey:         "level",
		TimeKey:          "ts",
		NameKey:          "name",
		CallerKey:        "caller",
		EncodeLevel:      zapcore.CapitalColorLevelEncoder,
		EncodeTime:       zapcore.ISO8601TimeEncoder,
		ConsoleSeparator: "\t",
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = lvl
	cfg.EncoderConfig = encoderCfg

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl
	Sugar = zl.Sugar()
	return nil
}

func LoggingHandler(h http.HandlerFunc) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqUri := r.RequestURI
		reqMethod := r.Method

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(&lw, r)

		reqDuration := time.Since(start)

		Sugar.Infoln(
			"reqURI:", reqUri,
			"reqMethod:", reqMethod,
			"reqDuration:", reqDuration,
			"respStatusCode:", responseData.status,
			"respSize:", responseData.size,
		)
	}

	return logFn
}
