package datautil

import (
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/c2pc/go-pkg/v2/utils/jsonutil"
	"github.com/jinzhu/copier"
)

// SliceSub возвращает элементы в срезе a, которые отсутствуют в срезе b (a - b).
func SliceSub[E comparable](a, b []E) []E {
	if len(b) == 0 {
		return a
	}
	k := make(map[E]struct{})
	for i := 0; i < len(b); i++ {
		k[b[i]] = struct{}{}
	}
	t := make(map[E]struct{})
	rs := make([]E, 0, len(a))
	for i := 0; i < len(a); i++ {
		e := a[i]
		if _, ok := t[e]; ok {
			continue
		}
		if _, ok := k[e]; ok {
			continue
		}
		rs = append(rs, e)
		t[e] = struct{}{}
	}
	return rs
}

// SliceSubAny возвращает элементы в срезе a, которые отсутствуют в срезе b (a - b).
// fn - функция, которая конвертирует элементы среза b в элементы, сравнимые с элементами среза a.
func SliceSubAny[E comparable, T any](a []E, b []T, fn func(t T) E) []E {
	return SliceSub(a, Slice(b, fn))
}

// SliceAnySub возвращает элементы в срезе a, которые отсутствуют в срезе b (a - b).
// fn - функция, которая извлекает сравнимое значение из элементов среза a.
func SliceAnySub[E any, T comparable](a, b []E, fn func(t E) T) []E {
	m := make(map[T]E)
	for i := 0; i < len(b); i++ {
		v := b[i]
		m[fn(v)] = v
	}
	var es []E
	for i := 0; i < len(a); i++ {
		v := a[i]
		if _, ok := m[fn(v)]; !ok {
			es = append(es, v)
		}
	}
	return es
}

// DistinctAny удаляет дубликаты из среза.
func DistinctAny[E any, K comparable](es []E, fn func(e E) K) []E {
	v := make([]E, 0, len(es))
	tmp := map[K]struct{}{}
	for i := 0; i < len(es); i++ {
		t := es[i]
		k := fn(t)
		if _, ok := tmp[k]; !ok {
			tmp[k] = struct{}{}
			v = append(v, t)
		}
	}
	return v
}

// DistinctAnyGetComparable возвращает уникальные значения из среза на основе функции fn.
func DistinctAnyGetComparable[E any, K comparable](es []E, fn func(e E) K) []K {
	v := make([]K, 0, len(es))
	tmp := map[K]struct{}{}
	for i := 0; i < len(es); i++ {
		t := es[i]
		k := fn(t)
		if _, ok := tmp[k]; !ok {
			tmp[k] = struct{}{}
			v = append(v, k)
		}
	}
	return v
}

// Distinct удаляет дубликаты из среза.
func Distinct[T comparable](ts []T) []T {
	if len(ts) < 2 {
		return ts
	} else if len(ts) == 2 {
		if ts[0] == ts[1] {
			return ts[:1]
		} else {
			return ts
		}
	}
	return DistinctAny(ts, func(t T) T {
		return t
	})
}

// Delete удаляет элементы из среза, поддерживает отрицательные числа для удаления элементов с конца.
func Delete[E any](es []E, index ...int) []E {
	switch len(index) {
	case 0:
		return es
	case 1:
		i := index[0]
		if i < 0 {
			i = len(es) + i
		}
		if len(es) <= i {
			return es
		}
		return append(es[:i], es[i+1:]...)
	default:
		tmp := make(map[int]struct{})
		for _, i := range index {
			if i < 0 {
				i = len(es) + i
			}
			tmp[i] = struct{}{}
		}
		v := make([]E, 0, len(es))
		for i := 0; i < len(es); i++ {
			if _, ok := tmp[i]; !ok {
				v = append(v, es[i])
			}
		}
		return v
	}
}

// DeleteAt удаляет элементы из среза, поддерживает отрицательные числа для удаления элемента с конца.
func DeleteAt[E any](es *[]E, index ...int) []E {
	v := Delete(*es, index...)
	*es = v
	return v
}

// IndexAny получает индекс элемента в срезе по сравнению с fn.
func IndexAny[E any, K comparable](e E, es []E, fn func(e E) K) int {
	k := fn(e)
	for i := 0; i < len(es); i++ {
		if fn(es[i]) == k {
			return i
		}
	}
	return -1
}

// IndexOf получает индекс элемента в срезе es.
func IndexOf[E comparable](e E, es ...E) int {
	return IndexAny(e, es, func(t E) E {
		return t
	})
}

// Contain проверяет, содержится ли элемент в срезе.
func Contain[E comparable](e E, es ...E) bool {
	return IndexOf(e, es...) >= 0
}

// DuplicateAny проверяет, есть ли дубликаты через fn.
func DuplicateAny[E any, K comparable](es []E, fn func(e E) K) bool {
	t := make(map[K]struct{})
	for _, e := range es {
		k := fn(e)
		if _, ok := t[k]; ok {
			return true
		}
		t[k] = struct{}{}
	}
	return false
}

// Duplicate проверяет, есть ли дубликаты в срезе.
func Duplicate[E comparable](es []E) bool {
	return DuplicateAny(es, func(e E) E {
		return e
	})
}

// SliceToMapOkAny преобразует срез в мапу.
func SliceToMapOkAny[E any, K comparable, V any](es []E, fn func(e E) (K, V, bool)) map[K]V {
	kv := make(map[K]V)
	for i := 0; i < len(es); i++ {
		t := es[i]
		if k, v, ok := fn(t); ok {
			kv[k] = v
		}
	}
	return kv
}

// SliceToMapAny преобразует срез в мапу.
func SliceToMapAny[E any, K comparable, V any](es []E, fn func(e E) (K, V)) map[K]V {
	return SliceToMapOkAny(es, func(e E) (K, V, bool) {
		k, v := fn(e)
		return k, v, true
	})
}

// SliceToMap преобразует срез в мапу.
func SliceToMap[E any, K comparable](es []E, fn func(e E) K) map[K]E {
	return SliceToMapOkAny(es, func(e E) (K, E, bool) {
		k := fn(e)
		return k, e, true
	})
}

// SliceSetAny преобразует срез в мапу[K]struct{}.
func SliceSetAny[E any, K comparable](es []E, fn func(e E) K) map[K]struct{} {
	return SliceToMapAny(es, func(e E) (K, struct{}) {
		return fn(e), struct{}{}
	})
}

// Filter фильтрует элементы с помощью заданной функции.
func Filter[E, T any](es []E, fn func(e E) (T, bool)) []T {
	rs := make([]T, 0, len(es))
	for i := 0; i < len(es); i++ {
		e := es[i]
		if t, ok := fn(e); ok {
			rs = append(rs, t)
		}
	}
	return rs
}

// Slice преобразует типы срезов батчами.
func Slice[E any, T any](es []E, fn func(e E) T) []T {
	v := make([]T, len(es))
	for i := 0; i < len(es); i++ {
		v[i] = fn(es[i])
	}
	return v
}

// SliceSet преобразует срез в мапу[E]struct{}.
func SliceSet[E comparable](es []E) map[E]struct{} {
	return SliceSetAny(es, func(e E) E {
		return e
	})
}

// HasKey проверяет, содержит ли мапа ключ.
func HasKey[K comparable, V any](m map[K]V, k K) bool {
	if m == nil {
		return false
	}
	_, ok := m[k]
	return ok
}

// Min возвращает минимальное значение из нескольких.
func Min[E Ordered](e ...E) E {
	v := e[0]
	for _, t := range e[1:] {
		if v > t {
			v = t
		}
	}
	return v
}

// Max возвращает максимальное значение из нескольких.
func Max[E Ordered](e ...E) E {
	v := e[0]
	for _, t := range e[1:] {
		if v < t {
			v = t
		}
	}
	return v
}

// Paginate выполняет пагинацию среза.
func Paginate[E any](es []E, pageNumber int, showNumber int) []E {
	if pageNumber <= 0 {
		return []E{}
	}
	if showNumber <= 0 {
		return []E{}
	}
	start := (pageNumber - 1) * showNumber
	end := start + showNumber
	if start >= len(es) {
		return []E{}
	}
	if end > len(es) {
		end = len(es)
	}
	return es[start:end]
}

// BothExistAny получает элементы, которые есть в нескольких срезах.
func BothExistAny[E any, K comparable](es [][]E, fn func(e E) K) []E {
	if len(es) == 0 {
		return []E{}
	}
	var idx int
	ei := make([]map[K]E, len(es))
	for i := 0; i < len(ei); i++ {
		e := es[i]
		if len(e) == 0 {
			return []E{}
		}
		kv := make(map[K]E)
		for j := 0; j < len(e); j++ {
			t := e[j]
			k := fn(t)
			kv[k] = t
		}
		ei[i] = kv
		if len(kv) < len(ei[idx]) {
			idx = i
		}
	}
	v := make([]E, 0, len(ei[idx]))
	for k := range ei[idx] {
		all := true
		for i := 0; i < len(ei); i++ {
			if i == idx {
				continue
			}
			if _, ok := ei[i][k]; !ok {
				all = false
				break
			}
		}
		if !all {
			continue
		}
		v = append(v, ei[idx][k])
	}
	return v
}

// BothExist находит общие элементы в срезах.
func BothExist[E comparable](es ...[]E) []E {
	return BothExistAny(es, func(e E) E {
		return e
	})
}

// Complete проверяет, равны ли a и b после удаления дубликатов (игнорируя порядок).
func Complete[E comparable](a []E, b []E) bool {
	return len(Single(a, b)) == 0
}

// Keys возвращает ключи из мапы.
func Keys[K comparable, V any](kv map[K]V) []K {
	ks := make([]K, 0, len(kv))
	for k := range kv {
		ks = append(ks, k)
	}
	return ks
}

// Values возвращает значения из мапы.
func Values[K comparable, V any](kv map[K]V) []V {
	vs := make([]V, 0, len(kv))
	for k := range kv {
		vs = append(vs, kv[k])
	}
	return vs
}

// Sort базовая сортировка типов.
func Sort[E Ordered](es []E, asc bool) []E {
	SortAny(es, func(a, b E) bool {
		if asc {
			return a < b
		} else {
			return a > b
		}
	})
	return es
}

// SortAny настраиваемый метод сортировки.
func SortAny[E any](es []E, fn func(a, b E) bool) {
	sort.Sort(&sortSlice[E]{
		ts: es,
		fn: fn,
	})
}

// If возвращает a, если true, иначе b.
func If[T any](isa bool, a, b T) T {
	if isa {
		return a
	}
	return b
}

// ToPtr возвращает указатель на t.
func ToPtr[T any](t T) *T {
	return &t
}

// Equal сравнивает два среза, включая порядок элементов.
func Equal[E comparable](a []E, b []E) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// EqualMaps Функция для сравнения карт
func EqualMaps(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || bv != v {
			return false
		}
	}
	return true
}

// Single возвращает элементы, которые присутствуют в a и отсутствуют в b, или наоборот.
func Single[E comparable](a, b []E) []E {
	kn := make(map[E]uint8)
	for _, e := range Distinct(a) {
		kn[e]++
	}
	for _, e := range Distinct(b) {
		kn[e]++
	}
	v := make([]E, 0, len(kn))
	for k, n := range kn {
		if n == 1 {
			v = append(v, k)
		}
	}
	return v
}

// Order сортирует ts по es.
func Order[E comparable, T any](es []E, ts []T, fn func(t T) E) []T {
	if len(es) == 0 || len(ts) == 0 {
		return ts
	}
	kv := make(map[E][]T)
	for i := 0; i < len(ts); i++ {
		t := ts[i]
		k := fn(t)
		kv[k] = append(kv[k], t)
	}
	rs := make([]T, 0, len(ts))
	for _, e := range es {
		vs := kv[e]
		delete(kv, e)
		rs = append(rs, vs...)
	}
	for k := range kv {
		rs = append(rs, kv[k]...)
	}
	return rs
}

// OrderPtr сортирует ts по es и обновляет указатель на ts.
func OrderPtr[E comparable, T any](es []E, ts *[]T, fn func(t T) E) []T {
	*ts = Order(es, *ts, fn)
	return *ts
}

// UniqueJoin объединяет уникальные строки в одну с помощью JSON.
func UniqueJoin(s ...string) string {
	data, _ := jsonutil.JsonMarshal(s)
	return string(data)
}

// sortSlice реализует интерфейс sort.Interface для кастомной сортировки.
type sortSlice[E any] struct {
	ts []E
	fn func(a, b E) bool
}

// Len возвращает длину среза.
func (o *sortSlice[E]) Len() int {
	return len(o.ts)
}

// Less сравнивает элементы для сортировки.
func (o *sortSlice[E]) Less(i, j int) bool {
	return o.fn(o.ts[i], o.ts[j])
}

// Swap меняет местами элементы в срезе.
func (o *sortSlice[E]) Swap(i, j int) {
	o.ts[i], o.ts[j] = o.ts[j], o.ts[i]
}

// Ordered определяет типы, которые могут быть отсортированы.
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64 | ~string
}

// NotNilReplace заменяет old на new_, если new_ не nil.
func NotNilReplace[T any](old, new_ *T) {
	if new_ == nil {
		return
	}
	*old = *new_
}

// StructFieldNotNilReplace копирует значения полей из src в dest, если они не нулевые.
func StructFieldNotNilReplace(dest, src any) {
	destVal := reflect.ValueOf(dest).Elem()
	srcVal := reflect.ValueOf(src).Elem()

	for i := 0; i < destVal.NumField(); i++ {
		destField := destVal.Field(i)
		srcField := srcVal.Field(i)

		// Check if the source field is valid
		if srcField.IsValid() {
			// Check if the target field can be set
			if destField.CanSet() {
				// Handling fields of slice type
				if destField.Kind() == reflect.Slice && srcField.Kind() == reflect.Slice {
					elemType := destField.Type().Elem()
					// Check if a slice element is a pointer to a structure
					if elemType.Kind() == reflect.Ptr && elemType.Elem().Kind() == reflect.Struct {
						// Create a new slice to store the copied elements
						newSlice := reflect.MakeSlice(destField.Type(), srcField.Len(), srcField.Cap())
						for j := 0; j < srcField.Len(); j++ {
							newElem := reflect.New(elemType.Elem())
							// Recursive update, retaining non-zero values
							StructFieldNotNilReplace(newElem.Interface(), srcField.Index(j).Interface())
							// Checks if the field of the new element is zero-valued, and if so, preserves the value at the corresponding position in the original slice
							for k := 0; k < newElem.Elem().NumField(); k++ {
								if newElem.Elem().Field(k).IsZero() {
									newElem.Elem().Field(k).Set(destField.Index(j).Elem().Field(k))
								}
							}
							newSlice.Index(j).Set(newElem)
						}
						destField.Set(newSlice)
					} else {
						destField.Set(srcField)
					}
				} else {
					// For non-sliced fields, update the source field if it is non-zero, otherwise keep the original value
					if !srcField.IsZero() {
						destField.Set(srcField)
					}
				}
			}
		}
	}
}

// Batch применяет функцию fn к каждому элементу ts.
func Batch[T any, V any](fn func(T) V, ts []T) []V {
	if ts == nil {
		return nil
	}
	res := make([]V, 0, len(ts))
	for i := range ts {
		res = append(res, fn(ts[i]))
	}
	return res
}

// InitSlice инициализирует срез, если он nil.
func InitSlice[T any](val *[]T) {
	if val != nil && *val == nil {
		*val = []T{}
	}
}

// InitMap инициализирует мапу, если она nil.
func InitMap[K comparable, V any](val *map[K]V) {
	if val != nil && *val == nil {
		*val = map[K]V{}
	}
}

// GetSwitchFromOptions извлекает значение ключа из настроек.
func GetSwitchFromOptions(Options map[string]bool, key string) (result bool) {
	if Options == nil {
		return true
	}
	if flag, ok := Options[key]; !ok || flag {
		return true
	}
	return false
}

// SetSwitchFromOptions устанавливает значение ключа в настройках.
func SetSwitchFromOptions(options map[string]bool, key string, value bool) {
	if options == nil {
		options = make(map[string]bool, 5)
	}
	options[key] = value
}

// CopyStructFields копирует поля из b в a.
func CopyStructFields(a any, b any, fields ...string) (err error) {
	return copier.Copy(a, b)
}

// GetElemByIndex возвращает элемент из массива по индексу.
func GetElemByIndex(array []int, index int) (int, error) {
	if index < 0 || index >= len(array) {
		return 0, errors.New(fmt.Sprintf("index out of range %s - %d %s - %v", "index", index, "array", array))
	}

	return array[index], nil
}
