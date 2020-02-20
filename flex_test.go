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

	record := ft.Build(map[string]interface{}{"name": "Adam"})
	require.True(t, record.Valid())

	assert.Equal(t, "Adam", record.Get("name"))
}

func TestTypeBuildError(t *testing.T) {
	ft := flex.Type{}
	ft.Field("age", flex.ConvertInt64())

	record := ft.Build(map[string]interface{}{"age": "abc"})
	require.False(t, record.Valid())
}

func TestRecordAttrs(t *testing.T) {
	ft := flex.Type{}
	ft.Field("a")
	ft.Field("b")
	ft.Field("c")
	ft.Field("d")

	record := ft.Build(map[string]interface{}{"a": "1", "b": "2", "c": "3"})
	assert.Equal(t, map[string]interface{}{"a": "1", "b": "2", "c": "3"}, record.Attrs())
}

func TestRecordPick(t *testing.T) {
	ft := flex.Type{}
	ft.Field("a")
	ft.Field("b")
	ft.Field("c")
	ft.Field("d")

	record := ft.Build(map[string]interface{}{"a": "1", "b": "2", "c": "3"})

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
		{nil, nil, true},
		{flex.UndefinedValue, flex.UndefinedValue, true},
	}

	for i, tt := range tests {
		value, err := flex.ConvertInt64().ConvertValue(tt.value)
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
		{flex.UndefinedValue, flex.UndefinedValue, true},
		{nil, nil, true},
		{"abc", nil, false},
	}

	for i, tt := range tests {
		value, err := flex.ConvertDecimal().ConvertValue(tt.value)
		assert.Equalf(t, tt.expected, value, "%d", i)
		assert.Equalf(t, tt.success, err == nil, "%d", i)
	}
}
