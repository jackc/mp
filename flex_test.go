package softstruct_test

import (
	"testing"

	"github.com/jackc/softstruct"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestType(t *testing.T) {
	ft := softstruct.Type{}
	ft.Field("name")

	record := ft.New(map[string]interface{}{"name": "Adam"})
	require.NoError(t, record.Errors())

	assert.Equal(t, "Adam", record.Get("name"))
}

func TestTypeNewError(t *testing.T) {
	ft := softstruct.Type{}
	ft.Field("age", softstruct.Int64())

	record := ft.New(map[string]interface{}{"age": "abc"})
	require.Error(t, record.Errors())
}

func TestTypeNewRequiredError(t *testing.T) {
	ft := softstruct.Type{}
	ft.Field("name", softstruct.Require())

	record := ft.New(map[string]interface{}{"misspelled": "adam"})
	require.Error(t, record.Errors())
}

func TestRecordAttrs(t *testing.T) {
	ft := softstruct.Type{}
	ft.Field("a")
	ft.Field("b")
	ft.Field("c")
	ft.Field("d")

	record := ft.New(map[string]interface{}{"a": "1", "b": "2", "c": "3"})
	assert.Equal(t, map[string]interface{}{"a": "1", "b": "2", "c": "3"}, record.Attrs())
}

func TestRecordGetPanicsWhenFieldNameNotInType(t *testing.T) {
	ft := softstruct.Type{}
	ft.Field("a")
	record := ft.New(map[string]interface{}{"b": "2"})
	assert.PanicsWithError(t, `"b" is not a field of type`, func() { record.Get("b") })
}

func TestRecordPick(t *testing.T) {
	ft := softstruct.Type{}
	ft.Field("a")
	ft.Field("b")
	ft.Field("c")
	ft.Field("d")

	record := ft.New(map[string]interface{}{"a": "1", "b": "2", "c": "3"})

	attrs := record.Pick("a", "b")
	assert.Equal(t, map[string]interface{}{"a": "1", "b": "2"}, attrs)

	attrs = record.Pick("c", "d")
	assert.Equal(t, map[string]interface{}{"c": "3"}, attrs)
}

func TestRecordPickPanicsWhenFieldNameNotInType(t *testing.T) {
	ft := softstruct.Type{}
	ft.Field("a")
	ft.Field("b")
	ft.Field("c")
	ft.Field("d")

	record := ft.New(map[string]interface{}{"a": "1", "b": "2", "c": "3"})

	assert.PanicsWithError(t, `"z" is not a field of type`, func() { record.Pick("a", "b", "z") })
}

func TestRequiredDefined(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		success  bool
	}{
		{"foo", "foo", true},
		{nil, nil, true},
		{softstruct.UndefinedValue, nil, false},
	}

	for i, tt := range tests {
		value, err := softstruct.RequireDefined().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestRequiredNotNil(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		success  bool
	}{
		{"foo", "foo", true},
		{nil, nil, false},
	}

	for i, tt := range tests {
		value, err := softstruct.RequireNotNil().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestRequire(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		success  bool
	}{
		{"foo", "foo", true},
		{"", nil, false},
		{nil, nil, false},
		{softstruct.UndefinedValue, nil, false},
	}

	for i, tt := range tests {
		value, err := softstruct.Require().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestInt64(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
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
		{softstruct.UndefinedValue, softstruct.UndefinedValue, true},
	}

	for i, tt := range tests {
		value, err := softstruct.Int64().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestFloat64(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		success  bool
	}{
		{1, float64(1), true},
		{"1", float64(1), true},
		{" 2 ", float64(2), true},
		{"10.5", float64(10.5), true},
		{"abc", nil, false},
		{nil, nil, true},
		{softstruct.UndefinedValue, softstruct.UndefinedValue, true},
	}

	for i, tt := range tests {
		value, err := softstruct.Float64().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestFloat32(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		success  bool
	}{
		{1, float32(1), true},
		{"1", float32(1), true},
		{" 2 ", float32(2), true},
		{"10.5", float32(10.5), true},
		{"abc", nil, false},
		{nil, nil, true},
		{softstruct.UndefinedValue, softstruct.UndefinedValue, true},
	}

	for i, tt := range tests {
		value, err := softstruct.Float32().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestBool(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
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
		{softstruct.UndefinedValue, softstruct.UndefinedValue, true},
	}

	for i, tt := range tests {
		value, err := softstruct.Bool().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestDecimal(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		success  bool
	}{
		{decimal.NewFromInt(1), decimal.NewFromInt(1), true},
		{1, decimal.NewFromInt(1), true},
		{"10.5", decimal.NewFromFloat(10.5), true},
		{" 7.7 ", decimal.NewFromFloat(7.7), true},
		{nil, nil, true},
		{softstruct.UndefinedValue, softstruct.UndefinedValue, true},
		{"abc", nil, false},
	}

	for i, tt := range tests {
		value, err := softstruct.Decimal().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestRecordSlice(t *testing.T) {
	softstructType := softstruct.NewType(func(tb softstruct.TypeBuilder) {
		tb.Field("n", softstruct.Int32(), softstruct.Require())
	})

	tests := []struct {
		value    interface{}
		expected interface{}
		success  bool
	}{
		{
			value:    []interface{}{map[string]interface{}{"n": 1}, map[string]interface{}{"n": 2}},
			expected: []*softstruct.Record{softstructType.New(map[string]interface{}{"n": 1}), softstructType.New(map[string]interface{}{"n": 2})},
			success:  true,
		},
		{
			value:    []interface{}{map[string]interface{}{"n": 1}, map[string]interface{}{"n": "abc"}},
			expected: nil,
			success:  false,
		},
		{value: softstruct.UndefinedValue, expected: softstruct.UndefinedValue, success: true},
		{value: nil, expected: nil, success: true},
		{[]int32{1, 2, 3}, nil, false},
		{[]interface{}{"1", "2", "3"}, nil, false},
		{[]interface{}{"1", 2, "3"}, nil, false},
		{"abc", nil, false},
		{42, nil, false},
	}

	for i, tt := range tests {
		value, err := softstruct.RecordSlice(softstructType).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestInt32Slice(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		success  bool
	}{
		{[]int32{1, 2, 3}, []int32{1, 2, 3}, true},
		{[]interface{}{"1", "2", "3"}, []int32{1, 2, 3}, true},
		{[]interface{}{"1", 2, "3"}, []int32{1, 2, 3}, true},
		{value: softstruct.UndefinedValue, expected: softstruct.UndefinedValue, success: true},
		{value: nil, expected: nil, success: true},
		{"abc", nil, false},
		{42, nil, false},
	}

	for i, tt := range tests {
		value, err := softstruct.Int32Slice().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestStringSlice(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		success  bool
	}{
		{[]string{"foo", "bar", "baz"}, []string{"foo", "bar", "baz"}, true},
		{[]interface{}{"foo", "bar", "baz"}, []string{"foo", "bar", "baz"}, true},
		{value: softstruct.UndefinedValue, expected: softstruct.UndefinedValue, success: true},
		{value: nil, expected: nil, success: true},
		{"abc", nil, false},
	}

	for i, tt := range tests {
		value, err := softstruct.StringSlice().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestTextField(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
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
		{value: softstruct.UndefinedValue, expected: softstruct.UndefinedValue, success: true},
		{value: nil, expected: nil, success: true},
	}

	for i, tt := range tests {
		value, err := softstruct.TextField().ConvertValue(tt.value)
		assert.Equalf(t, tt.success, err == nil, "%d: %s", i, tt.msg)
		assert.Equalf(t, tt.expected, value, "%d: %s", i, tt.msg)
	}
}

func TestNilifyEmptyString(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		success  bool
	}{
		{"foo", "foo", true},
		{"", nil, true},
		{softstruct.UndefinedValue, softstruct.UndefinedValue, true},
		{nil, nil, true},
	}

	for i, tt := range tests {
		value, err := softstruct.NilifyEmptyString().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestRequireStringMinLength(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		length   int
		success  bool
	}{
		{"foo", "foo", 1, true},
		{"f", "f", 1, true},
		{"", nil, 1, false},
		{1, nil, 1, false},
		{softstruct.UndefinedValue, nil, 1, false},
		{nil, nil, 1, false},
	}

	for i, tt := range tests {
		value, err := softstruct.RequireStringMinLength(tt.length).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestRequireStringMaxLength(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		length   int
		success  bool
	}{
		{"f", "f", 3, true},
		{"foo", "foo", 3, true},
		{"", "", 1, true},
		{1, nil, 1, false},
		{softstruct.UndefinedValue, nil, 1, false},
		{nil, nil, 1, false},
	}

	for i, tt := range tests {
		value, err := softstruct.RequireStringMaxLength(tt.length).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestRequireStringInclusion(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		success  bool
	}{
		{"foo", "foo", true},
		{"bar", nil, false},
		{"baz", "baz", true},
		{"", nil, false},
		{1, nil, false},
		{softstruct.UndefinedValue, nil, false},
		{nil, nil, false},
	}

	for i, tt := range tests {
		value, err := softstruct.RequireStringInclusion([]string{"foo", "baz", "abc"}).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestRequireDecimalLessThan(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		limit    decimal.Decimal
		success  bool
	}{
		{decimal.NewFromInt(1), decimal.NewFromInt(1), decimal.NewFromInt(10), true},
		{decimal.NewFromInt(10), nil, decimal.NewFromInt(10), false},
		{softstruct.UndefinedValue, nil, decimal.NewFromInt(10), false},
		{nil, nil, decimal.NewFromInt(10), false},
	}

	for i, tt := range tests {
		value, err := softstruct.RequireDecimalLessThan(tt.limit).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestRequireDecimalLessThanOrEqual(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		limit    decimal.Decimal
		success  bool
	}{
		{decimal.NewFromInt(1), decimal.NewFromInt(1), decimal.NewFromInt(10), true},
		{decimal.NewFromInt(10), decimal.NewFromInt(10), decimal.NewFromInt(10), true},
		{softstruct.UndefinedValue, nil, decimal.NewFromInt(10), false},
		{nil, nil, decimal.NewFromInt(10), false},
	}

	for i, tt := range tests {
		value, err := softstruct.RequireDecimalLessThanOrEqual(tt.limit).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestRequireDecimalGreaterThan(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		limit    decimal.Decimal
		success  bool
	}{
		{decimal.NewFromInt(11), decimal.NewFromInt(11), decimal.NewFromInt(10), true},
		{decimal.NewFromInt(10), nil, decimal.NewFromInt(10), false},
		{softstruct.UndefinedValue, nil, decimal.NewFromInt(10), false},
		{nil, nil, decimal.NewFromInt(10), false},
	}

	for i, tt := range tests {
		value, err := softstruct.RequireDecimalGreaterThan(tt.limit).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestRequireDecimalGreaterThanOrEqual(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		limit    decimal.Decimal
		success  bool
	}{
		{decimal.NewFromInt(11), decimal.NewFromInt(11), decimal.NewFromInt(10), true},
		{decimal.NewFromInt(10), decimal.NewFromInt(10), decimal.NewFromInt(10), true},
		{softstruct.UndefinedValue, nil, decimal.NewFromInt(10), false},
		{nil, nil, decimal.NewFromInt(10), false},
	}

	for i, tt := range tests {
		value, err := softstruct.RequireDecimalGreaterThanOrEqual(tt.limit).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestRequireInt64LessThan(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		limit    int64
		success  bool
	}{
		{int64(1), int64(1), 10, true},
		{int64(10), nil, 10, false},
		{softstruct.UndefinedValue, nil, 10, false},
		{nil, nil, 10, false},
	}

	for i, tt := range tests {
		value, err := softstruct.RequireInt64LessThan(tt.limit).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestRequireInt64LessThanOrEqual(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		limit    int64
		success  bool
	}{
		{int64(1), int64(1), 10, true},
		{int64(10), int64(10), 10, true},
		{softstruct.UndefinedValue, nil, 10, false},
		{nil, nil, 10, false},
	}

	for i, tt := range tests {
		value, err := softstruct.RequireInt64LessThanOrEqual(tt.limit).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestRequireInt64GreaterThan(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		limit    int64
		success  bool
	}{
		{int64(11), int64(11), 10, true},
		{int64(10), nil, 10, false},
		{softstruct.UndefinedValue, nil, 10, false},
		{nil, nil, 10, false},
	}

	for i, tt := range tests {
		value, err := softstruct.RequireInt64GreaterThan(tt.limit).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestRequireInt64GreaterThanOrEqual(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		limit    int64
		success  bool
	}{
		{int64(11), int64(11), 10, true},
		{int64(10), int64(10), 10, true},
		{softstruct.UndefinedValue, nil, 10, false},
		{nil, nil, 10, false},
	}

	for i, tt := range tests {
		value, err := softstruct.RequireInt64GreaterThanOrEqual(tt.limit).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func BenchmarkNewTypeAndRecord(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ft := softstruct.Type{}
		ft.Field("name", softstruct.String())
		ft.Field("age", softstruct.Int32())

		record := ft.New(map[string]interface{}{"name": "Adam", "age": 30})
		require.NoError(b, record.Errors())
	}
}

func BenchmarkRecord(b *testing.B) {
	ft := softstruct.Type{}
	ft.Field("name", softstruct.String())
	ft.Field("age", softstruct.Int32())

	for i := 0; i < b.N; i++ {
		record := ft.New(map[string]interface{}{"name": "Adam", "age": 30})
		require.NoError(b, record.Errors())
	}
}
