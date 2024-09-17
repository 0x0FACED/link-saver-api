package logger

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/0x0FACED/link-saver-api/config"
	"github.com/0x0FACED/link-saver-api/internal/wrap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	log *zap.Logger

	cfg config.LoggerConfig
}

func New(cfg config.LoggerConfig) *ZapLogger {
	dirName := "logs"
	err := os.MkdirAll(dirName, os.ModePerm)
	if err != nil {
		log.Fatalln("cant make dir: ", err)
		return nil
	}

	filename := time.Now().Format("2006-01-02") + ".log"
	filePath := filepath.Join(dirName, filename)

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln("cant open file: ", err)
		return nil
	}

	config := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	cEnc := zapcore.NewConsoleEncoder(config)
	fEnc := zapcore.NewConsoleEncoder(config)

	level, err := level(cfg.Level)
	if err != nil {
		level = 0 // default level is Info
	}

	core := zapcore.NewTee(
		zapcore.NewCore(cEnc, zapcore.AddSync(os.Stdout), zapcore.Level(level)),
		zapcore.NewCore(fEnc, zapcore.AddSync(file), zapcore.Level(level)),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))

	logger.Info("Logger successfully created!")

	return &ZapLogger{
		log: logger,
		cfg: cfg,
	}
}

func level(lvl string) (int8, error) {
	parsedInt, err := strconv.ParseInt(lvl, 10, 8) // 10 - основание, 8 - разрядность
	if err != nil {
		return -2, wrap.E("logger", "wrong level", err)
	}

	return int8(parsedInt), nil
}

func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("[2006-01-02 | 15:04:05]"))
}

func (z *ZapLogger) Info(wrappedMsg string, fields ...zap.Field) {
	z.log.Info("[MSG]: "+wrappedMsg, fields...)
}

func (z *ZapLogger) Debug(wrappedMsg string, fields ...zap.Field) {
	z.log.Debug("[MSG]: "+wrappedMsg, fields...)
}

func (z *ZapLogger) Error(wrappedMsg string, fields ...zap.Field) {
	z.log.Error("[MSG]: "+wrappedMsg, fields...)
}

func (z *ZapLogger) Fatal(wrappedMsg string, fields ...zap.Field) {
	z.log.Fatal("[MSG]: "+wrappedMsg, fields...)
}
