package redis

import (
	"context"
	"fmt"
	"log"

	"github.com/c2pc/go-pkg/v2/utils/constant"
	logger2 "github.com/c2pc/go-pkg/v2/utils/logger"
)

// logger - структура, использующая стандартный логгер для вывода логов
type logger struct {
	log *log.Logger
}

// defaultLogger создает и возвращает экземпляр logger с конфигурацией по умолчанию
func defaultLogger() logger {
	// Создание нового логгера с определенным идентификатором и конфигурацией
	writer := logger2.NewLogWriter(constant.REDIS_ID, false, 0)
	return logger{
		log: log.New(writer.Stdout, "\r\n\n", log.LstdFlags),
	}
}

// Printf выводит отформатированное сообщение в лог
func (l logger) Printf(ctx context.Context, format string, v ...interface{}) {
	// Форматирование сообщения и вывод его в лог
	_ = l.log.Output(2, fmt.Sprintf(format, v...))
}
