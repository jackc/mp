package softstruct

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/gofrs/uuid/v5"
	"github.com/shopspring/decimal"
)

type field struct {
	name       string
	converters []ValueConverter
}

type Type struct {
	fields map[string]*field
}

type TypeBuilder interface {
	Field(name string, converters ...ValueConverter)
}

func NewType(f func(tb TypeBuilder)) *Type {
	t := &Type{}
	f(t)
	return t
}

func (t *Type) Field(name string, converters ...ValueConverter) {
	if t.fields == nil {
		t.fields = make(map[string]*field)
	}

	t.fields[name] = &field{name: name, converters: converters}
}

func (t *Type) New(attrs map[string]any) *Record {
	r := &Record{
		t:         t,
		original:  attrs,
		converted: make(map[string]any, len(attrs)),
		errors:    make(map[string]error, len(attrs)),
	}

	for _, f := range t.fields {
		v := attrs[f.name]

		var err error
		for _, converter := range f.converters {
			v, err = converter.ConvertValue(v)
			if err != nil {
				break
			}
		}

		if err == nil {
			r.converted[f.name] = v
		} else {
			r.errors[f.name] = err
		}
	}

	return r
}

type ValueConverter interface {
	ConvertValue(any) (any, error)
}

type ValueConverterFunc func(any) (any, error)

func (vcf ValueConverterFunc) ConvertValue(v any) (any, error) {
	return vcf(v)
}

type Errors map[string]error

func (e Errors) Error() string {
	sb := &strings.Builder{}

	join := false
	for attr, err := range e {
		if join {
			sb.WriteString(", ")
		}
		fmt.Fprintf(sb, "%s %v", attr, err)
		join = true
	}

	return sb.String()
}

func (e Errors) MarshalJSON() ([]byte, error) {
	if len(e) == 0 {
		return []byte(`{}`), nil
	}

	m := make(map[string]any, len(e))
	for attr, err := range e {
		var val any
		if jm, ok := err.(json.Marshaler); ok {
			val = jm
		} else {
			val = err.Error()
		}
		m[attr] = val
	}

	return json.Marshal(m)
}

type sliceElementError struct {
	Index int
	Err   error
}

type sliceElementErrors []sliceElementError

func (e sliceElementErrors) Error() string {
	sb := &strings.Builder{}
	for i, ee := range e {
		if i > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(sb, "Element %d: %v", ee.Index, ee.Err)
	}
	return sb.String()
}

type Record struct {
	t         *Type
	original  map[string]any
	converted map[string]any
	errors    Errors
}

func (r *Record) Get(s string) any {
	if _, ok := r.t.fields[s]; !ok {
		panic(fmt.Errorf("%q is not a field of type", s))
	}

	return r.converted[s]
}

func (r *Record) Errors() error {
	if len(r.errors) == 0 {
		return nil
	}

	return r.errors
}

func (r *Record) Pick(keys ...string) map[string]any {
	m := make(map[string]any, len(keys))
	for _, k := range keys {
		if _, ok := r.t.fields[k]; !ok {
			panic(fmt.Errorf("%q is not a field of type", k))
		}

		if value, ok := r.converted[k]; ok {
			m[k] = value
		}
	}
	return m
}

func (r *Record) Attrs() map[string]any {
	return r.converted
}

func convertInt64(value any) (int64, error) {
	switch value := value.(type) {
	case int8:
		return int64(value), nil
	case uint8:
		return int64(value), nil
	case int16:
		return int64(value), nil
	case uint16:
		return int64(value), nil
	case int32:
		return int64(value), nil
	case uint32:
		return int64(value), nil
	case int64:
		return int64(value), nil
	case uint64:
		if value > math.MaxInt64 {
			return 0, errors.New("greater than maximum allowed number")
		}
		return int64(value), nil
	case int:
		if int64(value) < math.MinInt64 {
			return 0, errors.New("less than minimum allowed number")
		}
		if int64(value) > math.MaxInt64 {
			return 0, errors.New("greater than maximum allowed number")
		}
		return int64(value), nil
	case uint:
		if uint64(value) > math.MaxInt64 {
			return 0, errors.New("greater than maximum allowed number")
		}
		return int64(value), nil
	case float32:
		if value < math.MinInt64 {
			return 0, errors.New("less than minimum allowed number")
		}
		if value > math.MaxInt64 {
			return 0, errors.New("greater than maximum allowed number")
		}
		if float32(int64(value)) != value {
			return 0, errors.New("not a valid number")
		}
		return int64(value), nil
	case float64:
		if value < math.MinInt64 {
			return 0, errors.New("less than minimum allowed number")
		}
		if value > math.MaxInt64 {
			return 0, errors.New("greater than maximum allowed number")
		}
		if float64(int64(value)) != value {
			return 0, errors.New("not a valid number")
		}
		return int64(value), nil
	}

	s := fmt.Sprintf("%v", value)
	s = strings.TrimSpace(s)

	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, errors.New("not a valid number")
	}
	return num, nil
}

// Int64 returns a ValueConverter that converts value to an int64. If value is nil or a blank string nil is returned.
func Int64() ValueConverter {
	return ValueConverterFunc(func(value any) (any, error) {
		value = normalizeForParsing(value)

		if value == nil {
			return nil, nil
		}

		n, err := convertInt64(value)
		if err != nil {
			return nil, err
		}

		return n, nil
	})
}

func convertInt32(value any) (int32, error) {
	n, err := convertInt64(value)
	if err != nil {
		return 0, err
	}

	if n < math.MinInt32 {
		return 0, errors.New("less than minimum allowed number")
	}
	if n > math.MaxInt32 {
		return 0, errors.New("greater than maximum allowed number")
	}

	return int32(n), nil
}

// Int32 returns a ValueConverter that converts value to an int32. If value is nil or a blank string nil is returned.
func Int32() ValueConverter {
	return ValueConverterFunc(func(value any) (any, error) {
		value = normalizeForParsing(value)

		if value == nil {
			return nil, nil
		}

		n, err := convertInt32(value)
		if err != nil {
			return nil, err
		}

		return n, nil
	})
}

func convertFloat64(value any) (float64, error) {
	switch value := value.(type) {
	case int8:
		return float64(value), nil
	case uint8:
		return float64(value), nil
	case int16:
		return float64(value), nil
	case uint16:
		return float64(value), nil
	case int32:
		return float64(value), nil
	case uint32:
		return float64(value), nil
	case int64:
		return float64(value), nil
	case uint64:
		return float64(value), nil
	case int:
		return float64(value), nil
	case uint:
		return float64(value), nil
	case float32:
		return float64(value), nil
	case float64:
		return value, nil
	}

	s := fmt.Sprintf("%v", value)
	s = strings.TrimSpace(s)

	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, errors.New("not a valid number")
	}
	return num, nil
}

// Float64 returns a ValueConverter that converts value to an float64. If value is nil or a blank string nil is returned.
func Float64() ValueConverter {
	return ValueConverterFunc(func(value any) (any, error) {
		value = normalizeForParsing(value)

		if value == nil {
			return value, nil
		}

		n, err := convertFloat64(value)
		if err != nil {
			return nil, err
		}

		return n, nil
	})
}

func convertFloat32(value any) (float32, error) {
	n, err := convertFloat64(value)
	if err != nil {
		return 0, err
	}

	if n < -math.MaxFloat32 {
		return 0, errors.New("less than minimum allowed number")
	}
	if n > math.MaxFloat32 {
		return 0, errors.New("greater than maximum allowed number")
	}

	return float32(n), nil
}

// Float32 returns a ValueConverter that converts value to an float32. If value is nil or a blank string nil is
// returned.
func Float32() ValueConverter {
	return ValueConverterFunc(func(value any) (any, error) {
		value = normalizeForParsing(value)

		if value == nil {
			return value, nil
		}

		n, err := convertFloat32(value)
		if err != nil {
			return nil, err
		}

		return n, nil
	})
}

// Bool returns a ValueConverter that converts value to a bool. If value is nil or a blank string nil is returned.
func Bool() ValueConverter {
	return ValueConverterFunc(func(value any) (any, error) {
		value = normalizeForParsing(value)

		if value == nil {
			return nil, nil
		}

		switch value := value.(type) {
		case bool:
			return value, nil
		case string:
			value = strings.TrimSpace(value)
			b, err := strconv.ParseBool(value)
			if err != nil {
				return nil, err
			}
			return b, nil
		default:
			return nil, errors.New("not a valid boolean")
		}
	})
}

// UUID returns a ValueConverter that converts value to a uuid.UUID. If value is nil or a blank string nil is returned.
func UUID() ValueConverter {
	return ValueConverterFunc(func(value any) (any, error) {
		value = normalizeForParsing(value)

		if value == nil {
			return nil, nil
		}

		var uuidValue uuid.UUID
		var err error

		if value, ok := value.([]byte); ok {
			uuidValue, err = uuid.FromBytes(value)
			return uuidValue, err
		}

		s := fmt.Sprintf("%v", value)
		uuidValue, err = uuid.FromString(s)
		return uuidValue, err
	})
}

func convertDecimal(value any) (decimal.Decimal, error) {
	switch value := value.(type) {
	case decimal.Decimal:
		return value, nil
	case int64:
		return decimal.NewFromInt(value), nil
	case int:
		return decimal.NewFromInt(int64(value)), nil
	case int32:
		return decimal.NewFromInt32(value), nil
	case float32:
		return decimal.NewFromFloat32(value), nil
	case float64:
		return decimal.NewFromFloat(value), nil
	case string:
		value = strings.TrimSpace(value)
		return decimal.NewFromString(value)
	default:
		s := fmt.Sprintf("%v", value)
		s = strings.TrimSpace(s)
		return decimal.NewFromString(s)
	}
}

// Decimal returns a ValueConverter that converts value to a decimal.Decimal. If value is nil or a blank string nil is
// returned.
func Decimal() ValueConverter {
	return ValueConverterFunc(func(value any) (any, error) {
		value = normalizeForParsing(value)

		if value == nil {
			return nil, nil
		}

		n, err := convertDecimal(value)
		if err != nil {
			return nil, err
		}

		return n, nil
	})
}

func convertString(value any) string {
	switch value := value.(type) {
	case string:
		return value
	case []byte:
		return string(value)
	}

	return fmt.Sprint(value)
}

// String returns a ValueConverter that converts value to a string. If value is nil then nil is returned. It does not
// perform any normalization. In almost all cases, SingleLineString or MultiLineString should be used instead.
func String() ValueConverter {
	return ValueConverterFunc(func(value any) (any, error) {
		if value == nil {
			return value, nil
		}

		return convertString(value), nil
	})
}

func RecordSlice(t *Type) ValueConverter {
	return ValueConverterFunc(func(value any) (any, error) {
		if value == nil {
			return nil, nil
		}

		if value, ok := value.([]any); ok {
			var elErrs sliceElementErrors
			rs := make([]*Record, len(value))
			for i := range value {
				if r, ok := value[i].(map[string]any); ok {
					rs[i] = t.New(r)
					if rs[i].Errors() != nil {
						elErrs = append(elErrs, sliceElementError{Index: i, Err: rs[i].Errors()})
					}
				} else {
					return nil, fmt.Errorf("cannot convert to element %d to record", i)
				}
			}

			if elErrs != nil {
				return nil, elErrs
			}

			return rs, nil
		}

		return nil, errors.New("cannot convert to record slice")
	})
}

func Slice[T any](elementConverter ValueConverter) ValueConverter {
	return ValueConverterFunc(func(value any) (any, error) {
		if value == nil {
			return nil, nil
		}

		switch value := value.(type) {
		case []T:
			return value, nil
		case []any:
			ts := make([]T, len(value))
			var elErrs sliceElementErrors
			for i := range value {
				element, err := elementConverter.ConvertValue(value[i])
				if err != nil {
					elErrs = append(elErrs, sliceElementError{Index: i, Err: err})
				}
				if element, ok := element.(T); ok {
					ts[i] = element
				} else {
					elErrs = append(elErrs, sliceElementError{Index: i, Err: err})
				}
			}

			if elErrs != nil {
				return nil, elErrs
			}

			return ts, nil
		}

		return nil, fmt.Errorf("cannot convert to slice")
	})
}

// NotNil returns a ValueConverter that fails if value is nil.
func NotNil() ValueConverter {
	return ValueConverterFunc(func(value any) (any, error) {
		if value == nil {
			return nil, errors.New("cannot be nil")
		}
		return value, nil
	})
}

// Require returns a ValueConverter that returns an error if value is nil or "".
func Require() ValueConverter {
	return ValueConverterFunc(func(value any) (any, error) {
		if value == nil || value == "" {
			return nil, fmt.Errorf("cannot be nil or empty")
		}

		return value, nil
	})
}

func convertSlice(value any, converters []ValueConverter) (any, error) {
	v := value
	var err error

	for _, vc := range converters {
		v, err = vc.ConvertValue(v)
		if err != nil {
			break
		}
	}

	return v, err
}

func IfNotNil(converters ...ValueConverter) ValueConverter {
	return ValueConverterFunc(func(value any) (any, error) {
		if value == nil {
			return value, nil
		}

		return convertSlice(value, converters)
	})
}

// SingleLineString returns a ValueConverter that converts a string value to a normalized string. If value is nil then nil is
// returned. If value is not a string then an error is returned.
//
// It performs the following operations:
//   - Remove any invalid UTF-8
//   - Replace non-printable characters with standard space
//   - Remove spaces from left and right
func SingleLineString() ValueConverter {
	return ValueConverterFunc(func(value any) (any, error) {
		if value == nil {
			return nil, nil
		}

		if s, ok := value.(string); ok {
			s = strings.ToValidUTF8(s, "")
			s = strings.Map(func(r rune) rune {
				if unicode.IsPrint(r) {
					return r
				} else {
					return ' '
				}
			}, s)
			s = strings.TrimSpace(s)

			return s, nil
		}

		return nil, errors.New("not a string")
	})
}

// MultiLineString returns a ValueConverter that converts a string value to a normalized string. If value is nil then nil is
// returned. If value is not a string then an error is returned.
//
// It performs the following operations:
//   - Remove any invalid UTF-8
//   - Replace characters that are not graphic or space with standard space
func MultiLineString() ValueConverter {
	return ValueConverterFunc(func(value any) (any, error) {
		if value == nil {
			return nil, nil
		}

		if s, ok := value.(string); ok {
			s = strings.ToValidUTF8(s, "")
			s = strings.Map(func(r rune) rune {
				if unicode.IsGraphic(r) || unicode.IsSpace(r) {
					return r
				} else {
					return ' '
				}
			}, s)

			return s, nil
		}

		return nil, errors.New("not a string")
	})
}

// normalizeForParsing prepares value for parsing. If the value is not a string it is returned. Otherwise, space is
// trimmed from both sides of the string. If the string is now empty then nil is returned. Otherwise, the string is
// returned.
func normalizeForParsing(value any) any {
	if s, ok := value.(string); ok {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		return s
	}
	return value
}

// NilifyEmpty converts strings, slices, and maps where len(value) == 0 to nil. Any other value not modified.
func NilifyEmpty() ValueConverter {
	return ValueConverterFunc(func(value any) (any, error) {
		n, ok := tryLen(value)
		if ok && n == 0 {
			return nil, nil
		}
		return value, nil
	})
}

func requireStringTest(test func(string) bool, failErr error) ValueConverter {
	return ValueConverterFunc(func(value any) (any, error) {
		s, ok := value.(string)
		if !ok {
			return nil, errors.New("not a string")
		}

		if test(s) {
			return s, nil
		}

		return nil, failErr
	})
}

func tryLen(value any) (n int, ok bool) {
	s, ok := value.(string)
	if ok {
		return len(s), true
	}

	refval := reflect.ValueOf(value)
	switch refval.Kind() {
	case reflect.String, reflect.Slice, reflect.Map:
		return refval.Len(), true
	}

	return 0, false
}

// MinLen returns a ValueConverter that fails if len(value) < min. value must be a string, slice, or map. nil is
// returned unmodified.
func MinLen(min int) ValueConverter {
	return ValueConverterFunc(func(value any) (any, error) {
		if value == nil {
			return nil, nil
		}

		n, ok := tryLen(value)
		if !ok {
			return nil, errors.New("not a string, slice or map")
		}

		if n < min {
			return nil, fmt.Errorf("too short")
		}

		return value, nil
	})
}

// MaxLen returns a ValueConverter that fails if len(value) > max. value must be a string, slice, or map. nil is
// returned unmodified.
func MaxLen(max int) ValueConverter {
	return ValueConverterFunc(func(value any) (any, error) {
		if value == nil {
			return nil, nil
		}

		n, ok := tryLen(value)
		if !ok {
			return nil, errors.New("not a string, slice or map")
		}

		if n > max {
			return nil, fmt.Errorf("too long")
		}

		return value, nil
	})
}

// AllowStrings returns a ValueConverter that returns an error unless value is one of the allowedItems. If value is nil
// then nil is returned. If value is not a string then an error is returned.
func AllowStrings(allowedItems ...string) ValueConverter {
	set := make(map[string]struct{}, len(allowedItems))
	for _, item := range allowedItems {
		set[item] = struct{}{}
	}

	return ValueConverterFunc(func(value any) (any, error) {
		if value == nil {
			return value, nil
		}

		s, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("not allowed value")
		}

		if _, ok := set[s]; !ok {
			return nil, fmt.Errorf("not allowed value")
		}

		return value, nil
	})
}

// ExcludeStrings returns a ValueConverter that returns an error if value is one of the excludedItems. If value is nil
// then nil is returned. If value is not a string then an error is returned.
func ExcludeStrings(excludedItems ...string) ValueConverter {
	set := make(map[string]struct{}, len(excludedItems))
	for _, item := range excludedItems {
		set[item] = struct{}{}
	}

	return ValueConverterFunc(func(value any) (any, error) {
		if value == nil {
			return value, nil
		}

		s, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("not allowed value")
		}

		if _, ok := set[s]; ok {
			return nil, fmt.Errorf("not allowed value")
		}

		return value, nil
	})
}

func tryDecimal(value any) (n decimal.Decimal, ok bool) {
	var strValue string
	switch value := value.(type) {
	case decimal.Decimal:
		return value, true
	case int32:
		return decimal.NewFromInt32(value), true
	case int64:
		return decimal.NewFromInt(value), true
	case int:
		return decimal.NewFromInt(int64(value)), true
	case float32:
		return decimal.NewFromFloat32(value), true
	case float64:
		return decimal.NewFromFloat(value), true
	case string:
		strValue = value
	default:
		strValue = fmt.Sprint(value)
	}

	n, err := decimal.NewFromString(strValue)
	if err != nil {
		return decimal.Zero, false
	}

	return n, true
}

// LessThan returns a ValueConverter that fails unless value < x. x must be convertable to a decimal number or LessThan
// panics. value must be convertable to a decimal number. nil is returned unmodified.
func LessThan(x any) ValueConverter {
	dx, ok := tryDecimal(x)
	if !ok {
		panic(fmt.Errorf("%v is not convertable to a decimal number", x))
	}

	return ValueConverterFunc(func(value any) (any, error) {
		if value == nil {
			return nil, nil
		}

		n, ok := tryDecimal(value)
		if !ok {
			return nil, fmt.Errorf("not a number")
		}

		if !n.LessThan(dx) {
			return nil, fmt.Errorf("too large")
		}

		return value, nil
	})
}

// LessThanOrEqual returns a ValueConverter that fails unless value <= x. x must be convertable to a decimal number or
// LessThanOrEqual panics. value must be convertable to a decimal number. nil is returned unmodified.
func LessThanOrEqual(x any) ValueConverter {
	dx, ok := tryDecimal(x)
	if !ok {
		panic(fmt.Errorf("%v is not convertable to a decimal number", x))
	}

	return ValueConverterFunc(func(value any) (any, error) {
		if value == nil {
			return nil, nil
		}

		n, ok := tryDecimal(value)
		if !ok {
			return nil, fmt.Errorf("not a number")
		}

		if !n.LessThanOrEqual(dx) {
			return nil, fmt.Errorf("too large")
		}

		return value, nil
	})
}

// GreaterThan returns a ValueConverter that fails unless value > x. x must be convertable to a decimal number or
// GreaterThan panics. value must be convertable to a decimal number. nil is returned unmodified.
func GreaterThan(x any) ValueConverter {
	dx, ok := tryDecimal(x)
	if !ok {
		panic(fmt.Errorf("%v is not convertable to a decimal number", x))
	}

	return ValueConverterFunc(func(value any) (any, error) {
		if value == nil {
			return nil, nil
		}

		n, ok := tryDecimal(value)
		if !ok {
			return nil, fmt.Errorf("not a number")
		}

		if !n.GreaterThan(dx) {
			return nil, fmt.Errorf("too small")
		}

		return value, nil
	})
}

// GreaterThanOrEqual returns a ValueConverter that fails unless value >= x. x must be convertable to a decimal number
// or GreaterThanOrEqual panics. value must be convertable to a decimal number. nil is returned unmodified.
func GreaterThanOrEqual(x any) ValueConverter {
	dx, ok := tryDecimal(x)
	if !ok {
		panic(fmt.Errorf("%v is not convertable to a decimal number", x))
	}

	return ValueConverterFunc(func(value any) (any, error) {
		if value == nil {
			return nil, nil
		}

		n, ok := tryDecimal(value)
		if !ok {
			return nil, fmt.Errorf("not a number")
		}

		if !n.GreaterThanOrEqual(dx) {
			return nil, fmt.Errorf("too small")
		}

		return value, nil
	})
}
