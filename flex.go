package flex

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

type undefinedValueType string

const (
	UndefinedValue = undefinedValueType("undefined value")
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

func (t *Type) New(attrs map[string]interface{}) *Record {
	r := &Record{
		original:  attrs,
		converted: make(map[string]interface{}, len(attrs)),
		errors:    make(map[string]error, len(attrs)),
	}

	for _, f := range t.fields {
		v, present := attrs[f.name]
		if !present {
			v = UndefinedValue
		}

		var err error
		for _, converter := range f.converters {
			v, err = converter.ConvertValue(v)
			if err != nil {
				break
			}
		}

		if err == nil {
			if v != UndefinedValue {
				r.converted[f.name] = v
			}
		} else {
			r.errors[f.name] = err
		}
	}

	return r
}

type ValueConverter interface {
	ConvertValue(interface{}) (interface{}, error)
}

type ValueConverterFunc func(interface{}) (interface{}, error)

func (vcf ValueConverterFunc) ConvertValue(v interface{}) (interface{}, error) {
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

	m := make(map[string]interface{}, len(e))
	for attr, err := range e {
		var val interface{}
		if jm, ok := err.(json.Marshaler); ok {
			val = jm
		} else {
			val = err.Error()
		}
		m[attr] = val
	}

	return json.Marshal(m)
}

type Record struct {
	t         *Type
	original  map[string]interface{}
	converted map[string]interface{}
	errors    Errors
}

func (r *Record) Get(s string) interface{} {
	return r.converted[s]
}

func (r *Record) Errors() error {
	if len(r.errors) == 0 {
		return nil
	}

	return r.errors
}

func (r *Record) Pick(keys ...string) map[string]interface{} {
	m := make(map[string]interface{}, len(keys))
	for _, k := range keys {
		if value, ok := r.converted[k]; ok {
			m[k] = value
		}
	}
	return m
}

func (r *Record) Attrs() map[string]interface{} {
	return r.converted
}

func convertInt64(value interface{}) (int64, error) {
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
	}

	s := fmt.Sprintf("%v", value)
	s = strings.TrimSpace(s)

	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, errors.New("not a valid number")
	}
	return num, nil
}

// Int64 returns a ValueConverter that converts to an int64. Nil and UndefinedValue are returned unmodified.
func Int64() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		if value == nil || value == UndefinedValue {
			return value, nil
		}

		n, err := convertInt64(value)
		if err != nil {
			return nil, err
		}

		return n, nil
	})
}

func convertInt32(value interface{}) (int32, error) {
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

// Int32 returns a ValueConverter that converts to an int64. Nil and UndefinedValue are returned unmodified.
func Int32() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		if value == nil || value == UndefinedValue {
			return value, nil
		}

		n, err := convertInt32(value)
		if err != nil {
			return nil, err
		}

		return n, nil
	})
}

func convertFloat64(value interface{}) (float64, error) {
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

// Float64 returns a ValueConverter that converts to an int64. Nil and UndefinedValue are returned unmodified.
func Float64() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		if value == nil || value == UndefinedValue {
			return value, nil
		}

		n, err := convertFloat64(value)
		if err != nil {
			return nil, err
		}

		return n, nil
	})
}

func convertFloat32(value interface{}) (float32, error) {
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

// Float32 returns a ValueConverter that converts to an int64. Nil and UndefinedValue are returned unmodified.
func Float32() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		if value == nil || value == UndefinedValue {
			return value, nil
		}

		n, err := convertFloat32(value)
		if err != nil {
			return nil, err
		}

		return n, nil
	})
}

func Bool() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		if value == nil || value == UndefinedValue {
			return value, nil
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

func UUID() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		if value == nil || value == UndefinedValue {
			return value, nil
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

func convertDecimal(value interface{}) (decimal.Decimal, error) {
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

func Decimal() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		if value == nil || value == UndefinedValue {
			return value, nil
		}

		n, err := convertDecimal(value)
		if err != nil {
			return nil, err
		}

		return n, nil
	})
}

func convertString(value interface{}) string {
	switch value := value.(type) {
	case string:
		return value
	case []byte:
		return string(value)
	}

	return fmt.Sprintf("%v", value)
}

// String returns a ValueConverter that converts to a string. Nil and UndefinedValue are returned unmodified.
func String() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		if value == nil || value == UndefinedValue {
			return value, nil
		}

		return convertString(value), nil
	})
}

func StringSlice() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		if value == nil || value == UndefinedValue {
			return value, nil
		}

		switch value := value.(type) {
		case []string:
			return value, nil
		case []interface{}:
			ss := make([]string, len(value))
			for i := range value {
				ss[i] = convertString(value[i])
			}
			return ss, nil
		}

		return nil, errors.New("cannot convert to string slice")
	})
}

func Int32Slice() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		if value == nil || value == UndefinedValue {
			return value, nil
		}

		var err error

		switch value := value.(type) {
		case []int32:
			return value, nil
		case []interface{}:
			ns := make([]int32, len(value))
			for i := range value {
				ns[i], err = convertInt32(value[i])
				if err != nil {
					return nil, err
				}
			}
			return ns, nil
		}

		return nil, errors.New("cannot convert to int32 slice")
	})
}

func RequireDefined() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		if value == UndefinedValue {
			return nil, errors.New("must be defined")
		}
		return value, nil
	})
}

func RequireNotNil() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		if value == nil {
			return nil, errors.New("cannot be nil")
		}
		return value, nil
	})
}

func Require() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		valueConverters := []ValueConverter{
			RequireDefined(),
			RequireNotNil(),
		}

		for _, vc := range valueConverters {
			_, err := vc.ConvertValue(value)
			if err != nil {
				return nil, err
			}
		}

		if value == "" {
			return nil, errors.New("cannot be empty")
		}

		return value, nil
	})
}

func convertSlice(value interface{}, converters []ValueConverter) (interface{}, error) {
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

func IfDefined(converters ...ValueConverter) ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		if value == UndefinedValue {
			return value, nil
		}

		return convertSlice(value, converters)
	})
}

func IfNotNil(converters ...ValueConverter) ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		if value == nil {
			return value, nil
		}

		return convertSlice(value, converters)
	})
}

// TextField returns a ValueConverter that converts to a normalized string. Nil and UndefinedValue are returned
// unmodified.
//
// It performs the following operations:
// Remove any invalid UTF-8
// Replace non-printable characters with standard space
// Remove spaces from left and right
func TextField() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		if value == nil || value == UndefinedValue {
			return value, nil
		}

		if s, ok := value.(string); ok {
			return normalizeOneLineString(s), nil
		}
		return nil, errors.New("not a string")
	})
}

func normalizeOneLineString(s string) string {
	s = strings.ToValidUTF8(s, "")
	s = strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		} else {
			return ' '
		}
	}, s)
	s = strings.TrimSpace(s)

	return s
}

// NilifyEmptyString converts the empty string to nil. Any other value not modified.
func NilifyEmptyString() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		if value == "" {
			return nil, nil
		}
		return value, nil
	})
}

func requireStringTest(test func(string) bool, failErr error) ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
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

func RequireStringMinLength(n int) ValueConverter {
	return requireStringTest(func(s string) bool { return len(s) >= n }, errors.New("too short"))
}

func RequireStringMaxLength(n int) ValueConverter {
	return requireStringTest(func(s string) bool { return len(s) <= n }, errors.New("too long"))
}

func RequireStringInclusion(options []string) ValueConverter {
	return requireStringTest(
		func(s string) bool {
			for _, o := range options {
				if s == o {
					return true
				}
			}
			return false
		},
		errors.New("not allowed value"),
	)
}

func requireDecimalTest(test func(decimal.Decimal) bool, failErr error) ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		n, ok := value.(decimal.Decimal)
		if !ok {
			return nil, errors.New("not a decimal")
		}

		if test(n) {
			return n, nil
		}

		return nil, failErr
	})
}

func RequireDecimalLessThan(x decimal.Decimal) ValueConverter {
	return requireDecimalTest(func(n decimal.Decimal) bool { return n.LessThan(x) }, errors.New("too large"))
}

func RequireDecimalLessThanOrEqual(x decimal.Decimal) ValueConverter {
	return requireDecimalTest(func(n decimal.Decimal) bool { return n.LessThanOrEqual(x) }, errors.New("too large"))
}

func RequireDecimalGreaterThan(x decimal.Decimal) ValueConverter {
	return requireDecimalTest(func(n decimal.Decimal) bool { return n.GreaterThan(x) }, errors.New("too small"))
}

func RequireDecimalGreaterThanOrEqual(x decimal.Decimal) ValueConverter {
	return requireDecimalTest(func(n decimal.Decimal) bool { return n.GreaterThanOrEqual(x) }, errors.New("too small"))
}

func requireInt64Test(test func(int64) bool, failErr error) ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		n, ok := value.(int64)
		if !ok {
			return nil, errors.New("not a int64")
		}

		if test(n) {
			return n, nil
		}

		return nil, failErr
	})
}

func RequireInt64LessThan(x int64) ValueConverter {
	return requireInt64Test(func(n int64) bool { return n < x }, errors.New("too large"))
}

func RequireInt64LessThanOrEqual(x int64) ValueConverter {
	return requireInt64Test(func(n int64) bool { return n <= x }, errors.New("too large"))
}

func RequireInt64GreaterThan(x int64) ValueConverter {
	return requireInt64Test(func(n int64) bool { return n > x }, errors.New("too small"))
}

func RequireInt64GreaterThanOrEqual(x int64) ValueConverter {
	return requireInt64Test(func(n int64) bool { return n >= x }, errors.New("too small"))
}
