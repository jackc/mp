package mp_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/jackc/mp"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestType(t *testing.T) {
	ft := mp.NewType(
		mp.NewField("name"),
	)

	record := ft.Parse(map[string]any{"name": "Adam"})
	require.NoError(t, record.Errors())

	assert.Equal(t, "Adam", record.Get("name"))
}

func TestTypeNewError(t *testing.T) {
	ft := mp.NewType(
		mp.NewField("age", mp.Int64()),
	)

	record := ft.Parse(map[string]any{"age": "abc"})
	require.Error(t, record.Errors())
}

func TestTypeNewRequiredError(t *testing.T) {
	ft := mp.NewType(
		mp.NewField("name", mp.Require()),
	)

	record := ft.Parse(map[string]any{"misspelled": "adam"})
	require.Error(t, record.Errors())
}

func TestRecordAttrs(t *testing.T) {
	ft := mp.NewType(
		mp.NewField("a"),
		mp.NewField("b"),
		mp.NewField("c"),
		mp.NewField("d"))

	record := ft.Parse(map[string]any{"a": "1", "b": "2", "c": "3"})
	assert.Equal(t, map[string]any{"a": "1", "b": "2", "c": "3", "d": nil}, record.Attrs())
}

func TestRecordGetPanicsWhenFieldNameNotInType(t *testing.T) {
	ft := mp.NewType(
		mp.NewField("a"),
	)
	record := ft.Parse(map[string]any{"b": "2"})
	assert.PanicsWithError(t, `"b" is not a field of type`, func() { record.Get("b") })
}

func TestRecordPick(t *testing.T) {
	ft := mp.NewType(
		mp.NewField("a"),
		mp.NewField("b"),
		mp.NewField("c"),
		mp.NewField("d"),
	)

	record := ft.Parse(map[string]any{"a": "1", "b": "2", "c": "3"})

	attrs := record.Pick("a", "b")
	assert.Equal(t, map[string]any{"a": "1", "b": "2"}, attrs)

	attrs = record.Pick("c", "d")
	assert.Equal(t, map[string]any{"c": "3", "d": nil}, attrs)
}

func TestRecordPickPanicsWhenFieldNameNotInType(t *testing.T) {
	ft := mp.NewType(
		mp.NewField("a"),
		mp.NewField("b"),
		mp.NewField("c"),
		mp.NewField("d"),
	)

	record := ft.Parse(map[string]any{"a": "1", "b": "2", "c": "3"})

	assert.PanicsWithError(t, `"z" is not a field of type`, func() { record.Pick("a", "b", "z") })
}

func TestNotNil(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{"foo", "foo", true},
		{nil, nil, false},
	}

	for i, tt := range tests {
		value, err := mp.NotNil().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestRequire(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{"foo", "foo", true},
		{"", nil, false},
		{nil, nil, false},
	}

	for i, tt := range tests {
		value, err := mp.Require().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestInt64(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{1, int64(1), true},
		{"1", int64(1), true},
		{" 2 ", int64(2), true},
		{float32(12345678), int64(12345678), true},
		{float64(1234567890), int64(1234567890), true},
		{"10.5", nil, false},
		{"abc", nil, false},
		{nil, nil, true},
		{"", nil, true},
		{"  ", nil, true},
	}

	for i, tt := range tests {
		value, err := mp.Int64().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestFloat64(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{1, float64(1), true},
		{"1", float64(1), true},
		{" 2 ", float64(2), true},
		{"10.5", float64(10.5), true},
		{"abc", nil, false},
		{nil, nil, true},
		{"", nil, true},
		{"  ", nil, true},
	}

	for i, tt := range tests {
		value, err := mp.Float64().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestFloat32(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{1, float32(1), true},
		{"1", float32(1), true},
		{" 2 ", float32(2), true},
		{"10.5", float32(10.5), true},
		{"abc", nil, false},
		{nil, nil, true},
		{"", nil, true},
		{"  ", nil, true},
	}

	for i, tt := range tests {
		value, err := mp.Float32().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestBool(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{true, true, true},
		{false, false, true},
		{"true", true, true},
		{"t", true, true},
		{"false", false, true},
		{"f", false, true},
		{" true ", true, true},
		{"abc", nil, false},
		{nil, nil, true},
		{"", nil, true},
		{"  ", nil, true},
	}

	for i, tt := range tests {
		value, err := mp.Bool().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestTime(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{"foo", nil, false},
		{"2023-06-24", time.Date(2023, 6, 24, 0, 0, 0, 0, time.UTC), true},
		{"2023-06-24 20:41:50", time.Date(2023, 6, 24, 20, 41, 50, 0, time.UTC), true},
		{nil, nil, true},
		{"", nil, true},
		{"  ", nil, true},
	}

	for i, tt := range tests {
		value, err := mp.Time("2006-01-02", "2006-01-02 15:04:05").ConvertValue(tt.value)
		if tt.expected == nil {
			assert.Nilf(t, value, "%d", i)
		} else {
			expectedTime := tt.expected.(time.Time)
			valueTime, ok := value.(time.Time)
			assert.Truef(t, ok, "%d", i)
			assert.Truef(t, expectedTime.Equal(valueTime), "%d", i)
		}
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestDecimal(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{decimal.NewFromInt(1), decimal.NewFromInt(1), true},
		{1, decimal.NewFromInt(1), true},
		{"10.5", decimal.NewFromFloat(10.5), true},
		{" 7.7 ", decimal.NewFromFloat(7.7), true},
		{nil, nil, true},
		{"", nil, true},
		{"  ", nil, true},
		{"abc", nil, false},
	}

	for i, tt := range tests {
		value, err := mp.Decimal().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestSliceRecord(t *testing.T) {
	mpType := mp.NewType(
		mp.NewField("n", mp.Int32(), mp.Require()),
	)

	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{
			value:    []any{map[string]any{"n": 1}, map[string]any{"n": 2}},
			expected: []*mp.Record{mpType.Parse(map[string]any{"n": 1}), mpType.Parse(map[string]any{"n": 2})},
			success:  true,
		},
		{
			value:    []any{map[string]any{"n": 1}, map[string]any{"n": "abc"}},
			expected: nil,
			success:  false,
		},
		{value: nil, expected: nil, success: true},
		{[]int32{1, 2, 3}, nil, false},
		{[]any{"1", "2", "3"}, nil, false},
		{[]any{"1", 2, "3"}, nil, false},
		{"abc", nil, false},
		{42, nil, false},
	}

	for i, tt := range tests {
		value, err := mp.Slice[*mp.Record](mpType).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestSliceInt32(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{[]int32{1, 2, 3}, []int32{1, 2, 3}, true},
		{[]any{"1", "2", "3"}, []int32{1, 2, 3}, true},
		{[]any{"1", 2, "3"}, []int32{1, 2, 3}, true},
		{value: nil, expected: nil, success: true},
		{"abc", nil, false},
		{42, nil, false},
	}

	for i, tt := range tests {
		value, err := mp.Slice[int32](mp.Int32()).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestSliceString(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
	}{
		{[]string{"foo", "bar", "baz"}, []string{"foo", "bar", "baz"}, true},
		{[]any{"foo", "bar", "baz"}, []string{"foo", "bar", "baz"}, true},
		{value: nil, expected: nil, success: true},
		{"abc", nil, false},
	}

	for i, tt := range tests {
		value, err := mp.Slice[string](mp.SingleLineString()).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestSingleLineString(t *testing.T) {
	tests := []struct {
		value    any
		expected any
		success  bool
		msg      string
	}{
		{value: "a", expected: "a", success: true, msg: "no changes"},
		{value: " a", expected: "a", success: true, msg: "trim left"},
		{value: "a ", expected: "a", success: true, msg: "trim right"},
		{value: " a ", expected: "a", success: true, msg: "trim both sides"},
		{value: "a\xfe\xffa", expected: "aa", success: true, msg: "invalid UTF-8"},
		{value: "a\u200Ba", expected: "a a", success: true, msg: "replace non-normal spaces"},
		{value: "a\ta", expected: "a a", success: true, msg: "replace control character"},
		{value: "a\r\n", expected: "a", success: true, msg: "trim happens after replaced control character"},
		{value: nil, expected: nil, success: true},
	}

	for i, tt := range tests {
		value, err := mp.SingleLineString().ConvertValue(tt.value)
		assert.Equalf(t, tt.success, err == nil, "%d: %s", i, tt.msg)
		assert.Equalf(t, tt.expected, value, "%d: %s", i, tt.msg)
	}
}

func TestNilifyEmpty(t *testing.T) {
	type otherString string

	tests := []struct {
		value    any
		expected any
	}{
		{"foo", "foo"},
		{"", nil},
		{otherString(""), nil},
		{[]int{}, nil},
		{[]int{1}, []int{1}},
		{map[string]any{}, nil},
		{map[string]any{"foo": "bar"}, map[string]any{"foo": "bar"}},
		{nil, nil},
	}

	for i, tt := range tests {
		value, err := mp.NilifyEmpty().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.NoErrorf(t, err, "%d", i)
	}
}

func TestMinLen(t *testing.T) {
	tests := []struct {
		value      any
		expected   any
		length     int
		errMatcher *regexp.Regexp
	}{
		{"foo", "foo", 1, nil},
		{"f", "f", 1, nil},
		{"", nil, 1, regexp.MustCompile(`short`)},
		{1, nil, 1, regexp.MustCompile(`not a string`)},
		{[]int{1, 2, 3}, []int{1, 2, 3}, 1, nil},
		{[]int{}, nil, 1, regexp.MustCompile(`short`)},
		{map[string]any{}, nil, 1, regexp.MustCompile(`short`)},
		{map[string]any{"foo": "bar"}, map[string]any{"foo": "bar"}, 1, nil},
		{nil, nil, 1, nil},
	}

	for i, tt := range tests {
		value, err := mp.MinLen(tt.length).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		if tt.errMatcher == nil {
			require.NoError(t, err, "%d", i)
		} else {
			require.Regexpf(t, tt.errMatcher, err.Error(), "%d", i)
		}
	}
}

func TestMaxLen(t *testing.T) {
	tests := []struct {
		value      any
		expected   any
		length     int
		errMatcher *regexp.Regexp
	}{
		{"foo", "foo", 3, nil},
		{"f", "f", 3, nil},
		{"", "", 3, nil},
		{"abcd", nil, 3, regexp.MustCompile(`long`)},
		{1, nil, 3, regexp.MustCompile(`not a string`)},
		{[]int{1, 2, 3}, []int{1, 2, 3}, 3, nil},
		{[]int{1, 2, 3, 4}, nil, 3, regexp.MustCompile(`long`)},
		{map[string]any{"foo": "bar"}, map[string]any{"foo": "bar"}, 2, nil},
		{map[string]any{"foo": "bar", "baz": "quz"}, nil, 1, regexp.MustCompile(`long`)},
		{nil, nil, 1, nil},
	}

	for i, tt := range tests {
		value, err := mp.MaxLen(tt.length).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		if tt.errMatcher == nil {
			require.NoError(t, err, "%d", i)
		} else {
			require.Regexpf(t, tt.errMatcher, err.Error(), "%d", i)
		}
	}
}

func TestAllowStrings(t *testing.T) {
	tests := []struct {
		value         any
		allowedValues []string
		errMatcher    *regexp.Regexp
	}{
		{
			value:         "foo",
			allowedValues: []string{"foo", "bar"},
			errMatcher:    nil,
		},
		{
			value:         "quz",
			allowedValues: []string{"foo", "bar"},
			errMatcher:    regexp.MustCompile(`not allowed value`),
		},
	}

	for i, tt := range tests {
		value, err := mp.AllowStrings(tt.allowedValues...).ConvertValue(tt.value)
		if tt.errMatcher == nil {
			assert.Equalf(t, tt.value, value, "%d", i)
			assert.NoError(t, err, "%d", i)
		} else {
			assert.Regexpf(t, tt.errMatcher, err.Error(), "%d", i)
		}
	}
}

func TestExcludeStrings(t *testing.T) {
	tests := []struct {
		value          any
		excludedValues []string
		errMatcher     *regexp.Regexp
	}{
		{
			value:          "foo",
			excludedValues: []string{"foo", "bar"},
			errMatcher:     regexp.MustCompile(`not allowed value`),
		},
		{
			value:          "quz",
			excludedValues: []string{"foo", "bar"},
			errMatcher:     nil,
		},
	}

	for i, tt := range tests {
		value, err := mp.ExcludeStrings(tt.excludedValues...).ConvertValue(tt.value)
		if tt.errMatcher == nil {
			assert.Equalf(t, tt.value, value, "%d", i)
			assert.NoError(t, err, "%d", i)
		} else {
			assert.Regexpf(t, tt.errMatcher, err.Error(), "%d", i)
		}
	}
}

func TestLessThan(t *testing.T) {
	tests := []struct {
		value      any
		expected   any
		limit      any
		errMatcher *regexp.Regexp
	}{
		{decimal.NewFromInt(1), decimal.NewFromInt(1), decimal.NewFromInt(10), nil},
		{decimal.NewFromInt(10), nil, decimal.NewFromInt(10), regexp.MustCompile(`too large`)},
		{10, nil, 10, regexp.MustCompile(`too large`)},
		{32.5, nil, 10, regexp.MustCompile(`too large`)},
		{"11", nil, 10, regexp.MustCompile(`too large`)},
		{nil, nil, decimal.NewFromInt(10), nil},
	}

	for i, tt := range tests {
		value, err := mp.LessThan(tt.limit).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		if tt.errMatcher == nil {
			assert.NoError(t, err, "%d", i)
		} else {
			assert.Regexpf(t, tt.errMatcher, err.Error(), "%d", i)
		}
	}
}

func TestLessThanOrEqual(t *testing.T) {
	tests := []struct {
		value      any
		expected   any
		limit      any
		errMatcher *regexp.Regexp
	}{
		{decimal.NewFromInt(1), decimal.NewFromInt(1), decimal.NewFromInt(10), nil},
		{decimal.NewFromInt(10), decimal.NewFromInt(10), decimal.NewFromInt(10), nil},
		{decimal.NewFromInt(11), nil, decimal.NewFromInt(10), regexp.MustCompile(`too large`)},
		{10, 10, 10, nil},
		{32.5, nil, 10, regexp.MustCompile(`too large`)},
		{"11", nil, 10, regexp.MustCompile(`too large`)},
		{nil, nil, decimal.NewFromInt(10), nil},
	}

	for i, tt := range tests {
		value, err := mp.LessThanOrEqual(tt.limit).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		if tt.errMatcher == nil {
			assert.NoError(t, err, "%d", i)
		} else {
			assert.Regexpf(t, tt.errMatcher, err.Error(), "%d", i)
		}
	}
}

func TestGreaterThan(t *testing.T) {
	tests := []struct {
		value      any
		expected   any
		limit      any
		errMatcher *regexp.Regexp
	}{
		{decimal.NewFromInt(1), nil, decimal.NewFromInt(10), regexp.MustCompile(`too small`)},
		{decimal.NewFromInt(10), nil, decimal.NewFromInt(10), regexp.MustCompile(`too small`)},
		{decimal.NewFromInt(11), decimal.NewFromInt(11), decimal.NewFromInt(10), nil},
		{10, nil, 10, regexp.MustCompile(`too small`)},
		{32.5, 32.5, 10, nil},
		{"11", "11", 10, nil},
		{nil, nil, decimal.NewFromInt(10), nil},
	}

	for i, tt := range tests {
		value, err := mp.GreaterThan(tt.limit).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		if tt.errMatcher == nil {
			assert.NoError(t, err, "%d", i)
		} else {
			assert.Regexpf(t, tt.errMatcher, err.Error(), "%d", i)
		}
	}
}

func TestGreaterThanOrEqual(t *testing.T) {
	tests := []struct {
		value      any
		expected   any
		limit      any
		errMatcher *regexp.Regexp
	}{
		{decimal.NewFromInt(1), nil, decimal.NewFromInt(10), regexp.MustCompile(`too small`)},
		{decimal.NewFromInt(10), decimal.NewFromInt(10), decimal.NewFromInt(10), nil},
		{decimal.NewFromInt(11), decimal.NewFromInt(11), decimal.NewFromInt(10), nil},
		{10, 10, 10, nil},
		{32.5, 32.5, 10, nil},
		{"11", "11", 10, nil},
		{nil, nil, decimal.NewFromInt(10), nil},
	}

	for i, tt := range tests {
		value, err := mp.GreaterThanOrEqual(tt.limit).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		if tt.errMatcher == nil {
			assert.NoError(t, err, "%d", i)
		} else {
			assert.Regexpf(t, tt.errMatcher, err.Error(), "%d", i)
		}
	}
}

// func FieldType(field mp.Field) reflect.Type {
// 	var resultType reflect.Type
// 	for _, vc := range field.ValueConverters {
// 		if rt, ok := vc.(mp.ValueConverterResultTyper); ok {
// 			resultType = rt.ResultType()
// 		}
// 	}
// 	return resultType
// }

// func FieldRequired(field mp.Field) bool {
// 	return true
// }

// func TestIntrospectFieldType(t *testing.T) {
// 	field := mp.Field{
// 		ValueConverters: []mp.ValueConverter{mp.Int64()},
// 	}

// 	finalType := FieldType(field)
// 	require.Equal(t, reflect.TypeOf(int64(0)), finalType)
// }

func BenchmarkTypeParse(b *testing.B) {
	ft := mp.NewType(
		mp.NewField("name", mp.String()),
		mp.NewField("age", mp.Int32()),
	)

	for i := 0; i < b.N; i++ {
		record := ft.Parse(map[string]any{"name": "Adam", "age": 30})
		require.NoError(b, record.Errors())
	}
}
