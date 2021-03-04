package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
	"jefembe.co.tz/vas/smsgw/utils/propertymanager"
)

var Log *zap.SugaredLogger

func init() {
	//Encoder configs
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = TimestampEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	//lumberjack configs
	lumberjackLogger := &lumberjack.Logger{
		Filename:   propertymanager.GetStringProperty("logging.filename"),
		MaxSize:    propertymanager.GetIntProperty("logging.maxSize", 1500),
		MaxBackups: propertymanager.GetIntProperty("logging.maxBackups", 60),
		MaxAge:     propertymanager.GetIntProperty("logging.maxAge", 2),
		Compress:   true,
		LocalTime:  true,
	}
	writeSyncer := zapcore.AddSync(lumberjackLogger)
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)

	logger := zap.New(core, zap.AddCallerSkip(1), zap.AddCaller())
	Log = logger.Sugar()
}

func TimestampEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func Debug(template string, args ...interface{}) {
	Log.Debugf(template, args...)
}

func Info(template string, args ...interface{}) {
	Log.Infof(template, args...)
}

func Error(template string, args ...interface{}) {
	Log.Errorf(template, args...)
}

func Fatal(template string, args ...interface{}) {
	Log.Fatalf(template, args...)
}

func Warn(template string, args ...interface{}) {
	Log.Warnf(template, args...)
}

// func GetLoggerName(file string, line int, ok bool) string {
// 	var loggerName string
// 	if ok {
// 		packageName := strings.Split(file, "/")
// 		loggerName = fmt.Sprintf("%s/%s:%d ", packageName[len(packageName)-2], packageName[len(packageName)-1], line)
// 	}
// 	return loggerName
// }
