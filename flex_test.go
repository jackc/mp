package flex_test

import (
	"testing"

	"github.com/jackc/flex"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestType(t *testing.T) {
	ft := flex.Type{}
	ft.Field("name")

	record := ft.New(map[string]interface{}{"name": "Adam"})
	require.NoError(t, record.Errors())

	assert.Equal(t, "Adam", record.Get("name"))
}

func TestTypeNewError(t *testing.T) {
	ft := flex.Type{}
	ft.Field("age", flex.ConvertInt64())

	record := ft.New(map[string]interface{}{"age": "abc"})
	require.Error(t, record.Errors())
}

func TestTypeNewRequiredError(t *testing.T) {
	ft := flex.Type{}
	ft.Field("name", flex.Require())

	record := ft.New(map[string]interface{}{"misspelled": "adam"})
	require.Error(t, record.Errors())
}

func TestRecordAttrs(t *testing.T) {
	ft := flex.Type{}
	ft.Field("a")
	ft.Field("b")
	ft.Field("c")
	ft.Field("d")

	record := ft.New(map[string]interface{}{"a": "1", "b": "2", "c": "3"})
	assert.Equal(t, map[string]interface{}{"a": "1", "b": "2", "c": "3"}, record.Attrs())
}

func TestRecordPick(t *testing.T) {
	ft := flex.Type{}
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

func TestRequiredDefined(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		success  bool
	}{
		{"foo", "foo", true},
		{nil, nil, true},
		{flex.UndefinedValue, nil, false},
	}

	for i, tt := range tests {
		value, err := flex.RequireDefined().ConvertValue(tt.value)
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
		value, err := flex.RequireNotNil().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestConvertInt64(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		success  bool
	}{
		{1, int64(1), true},
		{"1", int64(1), true},
		{"10.5", nil, false},
		{"abc", nil, false},
		{nil, nil, false},
		{flex.UndefinedValue, nil, false},
	}

	for i, tt := range tests {
		value, err := flex.ConvertInt64().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestConvertBool(t *testing.T) {
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
		{"abc", nil, false},
		{nil, nil, false},
		{flex.UndefinedValue, nil, false},
	}

	for i, tt := range tests {
		value, err := flex.ConvertBool().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestConvertDecimal(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		success  bool
	}{
		{decimal.NewFromInt(1), decimal.NewFromInt(1), true},
		{1, decimal.NewFromInt(1), true},
		{"10.5", decimal.NewFromFloat(10.5), true},
		{flex.UndefinedValue, nil, false},
		{nil, nil, false},
		{"abc", nil, false},
	}

	for i, tt := range tests {
		value, err := flex.ConvertDecimal().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestConvertStringSlice(t *testing.T) {
	tests := []struct {
		value    interface{}
		expected interface{}
		success  bool
	}{
		{[]string{"foo", "bar", "baz"}, []string{"foo", "bar", "baz"}, true},
		{[]interface{}{"foo", "bar", "baz"}, []string{"foo", "bar", "baz"}, true},
		{flex.UndefinedValue, nil, false},
		{nil, nil, false},
		{"abc", nil, false},
	}

	for i, tt := range tests {
		value, err := flex.ConvertStringSlice().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func TestNormalizeTextField(t *testing.T) {
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
		{value: flex.UndefinedValue, expected: nil, success: false, msg: "undefined"},
		{value: nil, expected: nil, success: false, msg: "nil"},
	}

	for i, tt := range tests {
		value, err := flex.NormalizeTextField().ConvertValue(tt.value)
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
		{flex.UndefinedValue, flex.UndefinedValue, true},
		{nil, nil, true},
	}

	for i, tt := range tests {
		value, err := flex.NilifyEmptyString().ConvertValue(tt.value)
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
		{flex.UndefinedValue, nil, 1, false},
		{nil, nil, 1, false},
	}

	for i, tt := range tests {
		value, err := flex.RequireStringMinLength(tt.length).ConvertValue(tt.value)
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
		{flex.UndefinedValue, nil, 1, false},
		{nil, nil, 1, false},
	}

	for i, tt := range tests {
		value, err := flex.RequireStringMaxLength(tt.length).ConvertValue(tt.value)
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
		{flex.UndefinedValue, nil, decimal.NewFromInt(10), false},
		{nil, nil, decimal.NewFromInt(10), false},
	}

	for i, tt := range tests {
		value, err := flex.RequireDecimalLessThan(tt.limit).ConvertValue(tt.value)
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
		{flex.UndefinedValue, nil, decimal.NewFromInt(10), false},
		{nil, nil, decimal.NewFromInt(10), false},
	}

	for i, tt := range tests {
		value, err := flex.RequireDecimalLessThanOrEqual(tt.limit).ConvertValue(tt.value)
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
		{flex.UndefinedValue, nil, decimal.NewFromInt(10), false},
		{nil, nil, decimal.NewFromInt(10), false},
	}

	for i, tt := range tests {
		value, err := flex.RequireDecimalGreaterThan(tt.limit).ConvertValue(tt.value)
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
		{flex.UndefinedValue, nil, decimal.NewFromInt(10), false},
		{nil, nil, decimal.NewFromInt(10), false},
	}

	for i, tt := range tests {
		value, err := flex.RequireDecimalGreaterThanOrEqual(tt.limit).ConvertValue(tt.value)
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
		{flex.UndefinedValue, nil, 10, false},
		{nil, nil, 10, false},
	}

	for i, tt := range tests {
		value, err := flex.RequireInt64LessThan(tt.limit).ConvertValue(tt.value)
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
		{flex.UndefinedValue, nil, 10, false},
		{nil, nil, 10, false},
	}

	for i, tt := range tests {
		value, err := flex.RequireInt64LessThanOrEqual(tt.limit).ConvertValue(tt.value)
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
		{flex.UndefinedValue, nil, 10, false},
		{nil, nil, 10, false},
	}

	for i, tt := range tests {
		value, err := flex.RequireInt64GreaterThan(tt.limit).ConvertValue(tt.value)
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
		{flex.UndefinedValue, nil, 10, false},
		{nil, nil, 10, false},
	}

	for i, tt := range tests {
		value, err := flex.RequireInt64GreaterThanOrEqual(tt.limit).ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}

func BenchmarkNewTypeAndRecord(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ft := flex.Type{}
		ft.Field("name", flex.ConvertString())
		ft.Field("age", flex.ConvertInt32())

		record := ft.New(map[string]interface{}{"name": "Adam", "age": 30})
		require.NoError(b, record.Errors())
	}
}

func BenchmarkRecord(b *testing.B) {
	ft := flex.Type{}
	ft.Field("name", flex.ConvertString())
	ft.Field("age", flex.ConvertInt32())

	for i := 0; i < b.N; i++ {
		record := ft.New(map[string]interface{}{"name": "Adam", "age": 30})
		require.NoError(b, record.Errors())
	}
}
