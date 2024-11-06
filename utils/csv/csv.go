package csv

import (
	"encoding/csv"
	"errors"
	"io"
	"mime/multipart"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/apperr/code"
	"github.com/c2pc/go-pkg/v2/utils/translator"
	"github.com/jszwec/csvutil"
)

var (
	ErrReadingFile       = apperr.New("reading_file_error", apperr.WithTextTranslate(translator.Translate{translator.RU: "Ошибка чтения файла", translator.EN: "Error reading file"}), apperr.WithCode(code.InvalidArgument))
	ErrInvalidFileFormat = apperr.New("invalid_file_format", apperr.WithTextTranslate(translator.Translate{translator.RU: "Неверный формат файла", translator.EN: "Invalid file format"}), apperr.WithCode(code.InvalidArgument))
	ErrInvalidFieldCount = apperr.New("invalid_field_count", apperr.WithTextTranslate(translator.Translate{translator.RU: "Неверное количество полей", translator.EN: "Invalid field count"}), apperr.WithCode(code.InvalidArgument))
)

func UnMarshalCSVFromFile[T any](file *multipart.FileHeader) ([]T, error) {
	f, err := file.Open()
	if err != nil {
		return nil, ErrReadingFile.WithError(err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)

	dec, err := csvutil.NewDecoder(csvReader)
	if err != nil {
		return nil, ErrReadingFile.WithError(err)
	}

	var data []T
	for {
		var d T
		if err := dec.Decode(&d); err == io.EOF {
			break
		} else if err != nil {
			if errors.Is(err, csvutil.ErrFieldCount) {
				return nil, ErrInvalidFieldCount
			}

			return nil, ErrInvalidFileFormat.WithError(err)
		}
		data = append(data, d)
	}

	return data, nil
}
