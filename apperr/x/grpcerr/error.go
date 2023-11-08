package grpcerr

import (
	"github.com/c2pc/go-pkg/apperr"
	"github.com/c2pc/go-pkg/apperr/utils/appErrors"
	"github.com/c2pc/go-pkg/apperr/utils/code"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

func ParseError(err error) apperr.Apperr {
	var id string
	var text string
	st := status.Convert(err)
	grpcCode := code.GrpcToCode(st.Code())

	for _, detail := range st.Details() {
		switch t := detail.(type) {
		case *errdetails.BadRequest:
			for _, v := range t.GetFieldViolations() {
				switch v.GetField() {
				case "id":
					id = v.GetDescription()
				case "text":
					text = v.GetDescription()
				}
			}
		}
	}

	if grpcCode == code.Unavailable || grpcCode == code.DeadlineExceeded {
		return appErrors.ErrServerIsNotAvailable
	}

	return apperr.New(id, apperr.WithCode(grpcCode)).SetText(text)
}
