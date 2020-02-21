package flex

import (
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

func (t *Type) Field(name string, converters ...ValueConverter) {
	if t.fields == nil {
		t.fields = make(map[string]*field)
	}

	t.fields[name] = &field{name: name, converters: converters}
}

func (t *Type) Build(attrs map[string]interface{}) *Record {
	r := &Record{
		original:  attrs,
		converted: make(map[string]interface{}, len(attrs)),
		errors:    make(map[string]error, len(attrs)),
	}

	for _, f := range t.fields {
		if v, ok := attrs[f.name]; ok {
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

type Record struct {
	t         *Type
	original  map[string]interface{}
	converted map[string]interface{}
	errors    map[string]error
}

func (r *Record) Get(s string) interface{} {
	return r.converted[s]
}

func (r *Record) Valid() bool {
	return len(r.errors) == 0
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

	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, errors.New("not a valid number")
	}
	return num, nil
}

func ConvertInt64() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		n, err := convertInt64(value)
		if err != nil {
			return nil, err
		}

		return n, nil
	})
}

func ConvertInt32() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		n, err := convertInt64(value)
		if err != nil {
			return nil, err
		}

		if n < math.MinInt32 {
			return nil, errors.New("less than minimum allowed number")
		}
		if n > math.MaxInt32 {
			return nil, errors.New("greater than maximum allowed number")
		}

		return int32(n), nil
	})
}

func ConvertUUID() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
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
		return decimal.NewFromString(value)
	default:
		s := fmt.Sprintf("%v", value)
		return decimal.NewFromString(s)
	}
}

func ConvertDecimal() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		n, err := convertDecimal(value)
		if err != nil {
			return nil, err
		}

		return n, nil
	})
}

func ConvertString() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		switch value := value.(type) {
		case string:
			return value, nil
		case []byte:
			return string(value), nil
		}

		s := fmt.Sprintf("%v", value)
		return s, nil
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

		return value, nil
	})
}

func IfDefined(vc ValueConverter) ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		if value == UndefinedValue {
			return value, nil
		}
		return vc.ConvertValue(value)
	})
}

func IfNotNil(vc ValueConverter) ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		if value == nil {
			return value, nil
		}
		return vc.ConvertValue(value)
	})
}

// NormalizeTextField performed common normalization for a single line string. It performs the following operations:
//
// Remove any invalid UTF-8
// Replace non-printable characters with standard space
// Remove spaces from left and right
func NormalizeTextField() ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
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

func RequireStringLength(min, max int) ValueConverter {
	return ValueConverterFunc(func(value interface{}) (interface{}, error) {
		s, ok := value.(string)
		if !ok {
			return nil, errors.New("not a string")
		} else if len(s) < min {
			return nil, errors.New("is too short")
		} else if len(s) > max {
			return nil, errors.New("is too long")
		}
		return s, nil
	})
}
