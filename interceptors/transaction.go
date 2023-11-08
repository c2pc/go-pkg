package interceptors

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/c2pc/go-pkg/apperr"
	"github.com/c2pc/go-pkg/apperr/utils/appErrors"
	"github.com/c2pc/go-pkg/apperr/utils/translate"
	"github.com/c2pc/go-pkg/apperr/x/grpcerr"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

var (
	ErrInternalMethod = apperr.New("all",
		apperr.WithTitleTranslate(translate.Translate{translate.RU: "Ошибка"}),
		apperr.WithContext("all"),
	)

	ErrCommitDatabaseID = "commit_database_error"
	ErrPanicID          = "panic_error"
)

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

func (tr *Transaction) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, error error) {
	txHandle := tr.DB.
		WithContext(ctx).
		Begin(&sql.TxOptions{})
	defer func() {
		if r := recover(); r != nil {
			txHandle.Rollback()
			error = grpcerr.Response(ctx, ErrInternalMethod.WithError(appErrors.ErrInternal.NewID(ErrPanicID)))
			fmt.Println(r)
			return
		}
	}()

	newCtx := context.WithValue(ctx, "db_trx", txHandle)

	response, err := handler(newCtx, req)

	if err != nil {
		txHandle.Rollback()
	} else {
		if err := txHandle.Commit().Error; err != nil {
			return resp, grpcerr.Response(ctx, ErrInternalMethod.WithError(appErrors.ErrInternal.NewID(ErrCommitDatabaseID)))
		}
	}

	return response, err
}

func (tr *Transaction) StreamServerInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (error error) {
	txHandle := tr.DB.
		WithContext(stream.Context()).
		Begin(&sql.TxOptions{})
	defer func() {
		if r := recover(); r != nil {
			txHandle.Rollback()
			error = grpcerr.Response(stream.Context(), ErrInternalMethod.WithError(appErrors.ErrInternal.NewID(ErrPanicID)))
			fmt.Println(r)
			return
		}
	}()

	newCtx := context.WithValue(stream.Context(), "db_trx", txHandle)

	type serverStream struct {
		grpc.ServerStream
		ctx context.Context
	}

	stream = &serverStream{
		ServerStream: stream,
		ctx:          newCtx,
	}

	err := handler(srv, stream)

	if err != nil {
		txHandle.Rollback()
	} else {
		if err := txHandle.Commit().Error; err != nil {
			return grpcerr.Response(stream.Context(), ErrInternalMethod.WithError(appErrors.ErrInternal.NewID(ErrPanicID)))
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
