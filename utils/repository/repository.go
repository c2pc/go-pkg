package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/c2pc/go-pkg/v2/utils/apperr"
	"github.com/c2pc/go-pkg/v2/utils/clause"
	"github.com/c2pc/go-pkg/v2/utils/model"
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

type Repository[T any, C model.Model] interface {
	Trx(db *gorm.DB) T
	With(models ...string) T
	Joins(models ...string) T
	Omit(columns ...string) T
	Find(ctx context.Context, query string, args ...any) (*C, error)
	FindById(ctx context.Context, id int) (*C, error)
	Delete(ctx context.Context, query string, args ...any) error
	Create(ctx context.Context, u *C, returning ...string) (*C, error)
	Create2(ctx context.Context, u *[]C, returning ...string) (*[]C, error)
	CreateOrUpdate(ctx context.Context, u *C, onConflict []interface{}, doUpdates []interface{}, doCreates []interface{}, returning ...string) (*C, error)
	FirstOrCreate(ctx context.Context, u *C, returning string, query string, args ...any) (*C, error)
	Update(ctx context.Context, u *C, selects []interface{}, query string, args ...any) error
	Update2(ctx context.Context, u *[]C, selects []interface{}, query string, args ...any) error
	UpdateMap(ctx context.Context, u map[string]interface{}, query string, args ...any) error
	Count(ctx context.Context, query string, args ...any) (int64, error)
	List(ctx context.Context, f *model.Filter, query string, args ...any) ([]C, error)
	Paginate(ctx context.Context, p *model.Meta[C], query string, args ...any) error
	PluckIDs(ctx context.Context, query string, args ...any) ([]int, error)
	DB() *gorm.DB
}

type Repo[C model.Model] struct {
	searchable   clause.FieldSearchable
	fieldOrderBy clause.FieldOrderBy
	db           *gorm.DB
	with         []string
}

func NewRepository[C model.Model](db *gorm.DB, fieldSearchable clause.FieldSearchable, fieldOrderBy clause.FieldOrderBy) Repo[C] {
	return Repo[C]{
		searchable:   fieldSearchable,
		fieldOrderBy: fieldOrderBy,
		db:           db,
	}
}

func (r Repo[C]) Model() C {
	m := new(C)
	return *m
}

func (r Repo[C]) Trx(db *gorm.DB) Repo[C] {
	if db != nil {
		r.db = db
	}
	return r
}

func (r Repo[C]) DB() *gorm.DB {
	return r.db.Session(&gorm.Session{})
}

func (r Repo[C]) SetDB(db *gorm.DB) Repo[C] {
	r.db = db
	return r
}

func (r Repo[C]) FieldSearchable() clause.FieldSearchable {
	return r.searchable
}

func (r Repo[C]) FieldOrderBy() clause.FieldOrderBy {
	return r.fieldOrderBy
}

func (r Repo[C]) Exists(ctx context.Context, field string, value interface{}, excludeField string, excludeValue interface{}) error {
	row := r.Model()
	res := r.DB().
		WithContext(ctx).
		Select(field).
		Scopes(func(tx *gorm.DB) *gorm.DB {
			if excludeField == "" {
				return tx
			}
			return tx.Where(excludeField+" <> ?", excludeValue)
		}).
		Where(field+" = ?", value).
		First(&row)
	if err := res.Error; err != nil {
		return r.Error(ctx, err)
	}

	return nil
}

func (r Repo[C]) Error(ctx context.Context, err error) error {
	var appError apperr.Error
	if errors.As(err, &appError) {
		return err
	}

	var pgError = &pgconn.PgError{}
	var mysqlErr *mysql.MySQLError

	switch {
	case apperr.Is(err, gorm.ErrRecordNotFound):
		return apperr.ErrDBRecordNotFound.WithError(err)
	case errors.As(err, &pgError):
		if pgError.Code == "23505" {
			return apperr.ErrDBDuplicated.WithError(err)
		}
	case errors.As(err, &mysqlErr):
		if mysqlErr.Number == 1062 {
			return apperr.ErrDBDuplicated.WithError(err)
		}
	case ctx.Err() != nil:
		return apperr.ErrContextCanceled.WithError(err)
	}

	return apperr.ErrDBInternal.WithError(err)
}

func (r Repo[C]) With(models ...string) Repo[C] {
	if len(models) > 0 {
		newModels := r.reformatModels(models...)

		for _, m := range newModels {
			r.with = append(r.with, m)
			if strings.Index(m, ".") != -1 || m[len(m)-1:] == "s" {
				r.db = r.db.Preload(m)
			} else {
				r.db = r.db.Joins(m)
			}
		}
	}

	return r
}

func (r Repo[C]) Omit(columns ...string) Repo[C] {
	if len(columns) > 0 {
		r.db = r.db.Omit(columns...)
	}

	return r
}

func (r Repo[C]) Joins(models ...string) Repo[C] {
	newModels := r.reformatModels(models...)
	for _, m := range newModels {
		if isJoin := func() bool {
			if m[len(m)-1:] == "s" {
				return false
			}

			for _, j := range r.with {
				if j == m {
					return true
				}
			}

			for _, j := range r.db.Statement.Joins {
				if j.Name == m {
					return true
				}
			}
			return false
		}(); !isJoin {
			r.db = r.db.Joins(m)
		}
	}

	return r
}

func (r Repo[C]) reformatModels(models ...string) []string {
	if len(models) == 0 {
		return []string{}
	}
	var newModels []string
	for _, m := range models {
		if m != "" {
			if strings.Index(m, "JOIN") != -1 {
				newModels = append(newModels, m)
			} else {
				m = strings.ReplaceAll(cases.Title(language.English).String(strings.ReplaceAll(m, ".", " ")), " ", ".")
				reg := regexp.MustCompile(`_(.)`)
				m = reg.ReplaceAllStringFunc(m, func(w string) string {
					return strings.ToUpper(w[1:])
				})
				newModels = append(newModels, m)
			}
		}
	}

	return newModels
}

func (r Repo[C]) QuoteTo(str string) string {
	quote := `"`
	switch r.DB().Dialector.Name() {
	case "postgres":
		quote = `"`
	case "mysql":
		quote = "`"
	}

	return strings.ReplaceAll(str, `"`, quote)
}

func (r Repo[C]) Find(ctx context.Context, query string, args ...any) (*C, error) {
	row := r.Model()

	res := r.
		DB().
		WithContext(ctx).
		Scopes(clause.Where(r.QuoteTo, query, args...)).
		First(&row)
	if err := res.Error; err != nil {
		return nil, r.Error(ctx, err)
	}

	return &row, nil
}

func (r Repo[C]) FindById(ctx context.Context, id int) (*C, error) {
	row := r.Model()
	query := fmt.Sprintf(`%s."id" = ?`, r.Model().TableName())
	res := r.
		DB().
		WithContext(ctx).
		Scopes(clause.Where(r.QuoteTo, query, id)).
		First(&row)
	if err := res.Error; err != nil {
		return nil, r.Error(ctx, err)
	}

	return &row, nil
}

func (r Repo[C]) Delete(ctx context.Context, query string, args ...any) error {
	row := r.Model()

	res := r.
		DB().
		WithContext(ctx).
		Scopes(clause.Where(r.QuoteTo, query, args...)).
		Delete(&row)
	if err := res.Error; err != nil {
		return r.Error(ctx, err)
	}

	return nil
}

func (r Repo[C]) Create(ctx context.Context, u *C, returning ...string) (*C, error) {
	res := r.DB().
		WithContext(ctx).
		Clauses(clause.Returning(returning...)).
		Create(u)
	if err := res.Error; err != nil {
		return nil, r.Error(ctx, err)
	}

	return u, nil
}

func (r Repo[C]) Create2(ctx context.Context, u *[]C, returning ...string) (*[]C, error) {
	res := r.DB().
		WithContext(ctx).
		Clauses(clause.Returning(returning...)).
		Create(u)
	if err := res.Error; err != nil {
		return nil, r.Error(ctx, err)
	}

	return u, nil
}

func (r Repo[C]) CreateOrUpdate(ctx context.Context, u *C, onConflict []interface{}, doUpdates []interface{}, doCreates []interface{}, returning ...string) (*C, error) {
	res := r.DB().
		WithContext(ctx).
		Clauses(clause.Returning(returning...))

	var selected []interface{}
	if onConflict != nil && len(onConflict) > 0 {
		selected = append(selected, onConflict...)
	}

	if doUpdates != nil && len(doUpdates) > 0 {
		selected = append(selected, doUpdates...)
	}

	if doCreates != nil && len(doCreates) > 0 {
		selected = append(selected, doCreates...)
	}

	if onConflict != nil && len(onConflict) > 0 && doUpdates != nil && len(doUpdates) > 0 {
		res = res.
			Clauses(clause.OnConflict(onConflict, doUpdates))
	}

	res = res.
		Select(selected[0], selected[0:]...).
		Create(u)
	if err := res.Error; err != nil {
		return nil, r.Error(ctx, err)
	}

	return u, nil
}

func (r Repo[C]) FirstOrCreate(ctx context.Context, u *C, returning string, query string, args ...any) (*C, error) {
	res := r.DB().
		WithContext(ctx).
		Scopes(clause.Where(r.QuoteTo, query, args...)).
		Clauses(clause.Returning(returning)).
		FirstOrCreate(u)
	if err := res.Error; err != nil {
		return nil, r.Error(ctx, err)
	}

	return u, nil
}

func (r Repo[C]) Update(ctx context.Context, u *C, selects []interface{}, query string, args ...any) error {
	res := r.
		DB().
		WithContext(ctx).
		Model(r.Model()).
		Scopes(clause.Where(r.QuoteTo, query, args...))

	if selects != nil && len(selects) > 0 {
		res = res.Select(selects[0], selects...)
	}

	res = res.Updates(u)
	if err := res.Error; err != nil {
		return r.Error(ctx, err)
	}

	return nil
}

func (r Repo[C]) Update2(ctx context.Context, u *[]C, selects []interface{}, query string, args ...any) error {
	res := r.
		DB().
		WithContext(ctx).
		Model(r.Model()).
		Scopes(clause.Where(r.QuoteTo, query, args...))

	if selects != nil && len(selects) > 0 {
		res = res.Select(selects[0], selects...)
	}

	res = res.Updates(u)
	if err := res.Error; err != nil {
		return r.Error(ctx, err)
	}

	return nil
}

func (r Repo[C]) UpdateMap(ctx context.Context, u map[string]interface{}, query string, args ...any) error {
	res := r.
		DB().
		WithContext(ctx).
		Model(r.Model()).
		Scopes(clause.Where(r.QuoteTo, query, args...)).
		Updates(u)
	if err := res.Error; err != nil {
		return r.Error(ctx, err)
	}

	return nil
}

func (r Repo[C]) Count(ctx context.Context, query string, args ...any) (int64, error) {
	var count int64
	res := r.
		DB().
		WithContext(ctx).
		Model(r.Model()).
		Scopes(clause.Where(r.QuoteTo, query, args...)).
		Count(&count)
	if err := res.Error; err != nil {
		return 0, r.Error(ctx, err)
	}

	return count, nil
}

func (r Repo[C]) List(ctx context.Context, f *model.Filter, query string, args ...any) ([]C, error) {
	repo, err := r.Where(f.Where)
	if err != nil {
		return nil, r.Error(ctx, err)
	}

	repo, err = repo.OrderBy(f.OrderBy)
	if err != nil {
		return nil, r.Error(ctx, err)
	}

	var rows []C
	res := repo.
		DB().
		WithContext(ctx).
		Model(r.Model()).
		Scopes(clause.Where(r.QuoteTo, query, args...)).
		Find(&rows)
	if err := res.Error; err != nil {
		return nil, r.Error(ctx, err)
	}

	return rows, nil
}

func (r Repo[C]) PluckIDs(ctx context.Context, query string, args ...any) ([]int, error) {
	var rows []int
	res := r.
		DB().
		WithContext(ctx).
		Model(r.Model()).
		Scopes(clause.Where(r.QuoteTo, query, args...)).
		Pluck("id", &rows)
	if err := res.Error; err != nil {
		return nil, r.Error(ctx, err)
	}

	return rows, nil
}

func (r Repo[C]) Paginate(ctx context.Context, p *model.Meta[C], query string, args ...any) error {
	repo, err := r.Where(p.Where)
	if err != nil {
		return r.Error(ctx, err)
	}

	db := repo.
		DB().
		WithContext(ctx).
		Model(r.Model()).
		Scopes(clause.Where(r.QuoteTo, query, args...))

	if p.MustReturnTotalRows {
		res := db.Session(&gorm.Session{}).Count(&p.TotalRows)
		if err := res.Error; err != nil {
			return r.Error(ctx, err)
		}
	}

	repo, err = repo.SetDB(db).OrderBy(p.OrderBy)
	if err != nil {
		return r.Error(ctx, err)
	}

	var rows []C
	res2 := repo.
		DB().
		Scopes(clause.Limit(p.GetLimit(), p.GetOffset())).
		Find(&rows)
	if err := res2.Error; err != nil {
		return r.Error(ctx, err)
	}
	p.Rows = rows

	return nil
}

func (r Repo[C]) OrderBy(orderBy map[string]string) (Repo[C], error) {
	if len(orderBy) > 0 {
		query, joins, err := clause.OrderByFilter(r.QuoteTo, orderBy, r.FieldOrderBy())
		if err != nil {
			return r, err
		}

		r.db = r.Joins(joins...).DB().Order(query)
	}
	return r, nil
}

func (r Repo[C]) Where(where *clause.ExpressionWhere) (Repo[C], error) {
	if where != nil {
		query, args, joins, err := clause.WhereFilter(r.QuoteTo, where, r.FieldSearchable())
		if err != nil {
			return r, err
		}

		r.db = r.Joins(joins...).DB().Where(query, args...)
	}
	return r, nil
}
