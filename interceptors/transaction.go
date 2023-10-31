package interceptors

import (
	"context"
	"database/sql"
	"github.com/c2pc/go-pkg/apperr"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

var ErrInternal = apperr.NewMethod("", "all").
	WithTitleTranslate(apperr.Translate{"ru": "Ошибка"})

type ITransaction interface {
	DBTransactionMiddleware() gin.HandlerFunc
}

type Transaction struct {
	DB *gorm.DB
}

func NewTr(db *gorm.DB) *Transaction {
	return &Transaction{
		DB: db,
	}
}

func (tr *Transaction) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	txHandle := tr.DB.
		WithContext(ctx).
		Begin(&sql.TxOptions{})
	defer func() {
		if r := recover(); r != nil {
			txHandle.Rollback()
			panic(r)
			return
		}
	}()

	newCtx := context.WithValue(ctx, "db_trx", txHandle)

	resp, err := handler(newCtx, req)

	if err != nil {
		txHandle.Rollback()
	} else {
		if err := txHandle.Commit().Error; err != nil {
			return resp, apperr.GRPCResponse(ErrInternal.Combine(apperr.ErrInternal.WithError(err)))
		}
	}

	return resp, err
}

func (tr *Transaction) StreamServerInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	txHandle := tr.DB.
		WithContext(stream.Context()).
		Begin(&sql.TxOptions{})
	defer func() {
		if r := recover(); r != nil {
			txHandle.Rollback()
			panic(r)
			return
		}
	}()

	newCtx := context.WithValue(stream.Context(), "db_trx", txHandle)

	type serverStream struct {
		grpc.ServerStream
		ctx context.Context
	}

	err := handler(srv, &serverStream{
		ServerStream: stream,
		ctx:          newCtx,
	})

	if err != nil {
		txHandle.Rollback()
	} else {
		if err := txHandle.Commit().Error; err != nil {
			return apperr.GRPCResponse(ErrInternal.Combine(apperr.ErrInternal.WithError(err)))
		}
	}

	return err
}

func TxHandle(ctx context.Context) *gorm.DB {
	txHandle, ok := ctx.Value("db_trx").(*gorm.DB)
	if !ok {
		return nil
	}

	return txHandle
}
