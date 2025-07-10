package extract

import (
	"testing"
)

func TestString(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var pointer *string
		want := ""

		if got := String(pointer); got != want {
			t.Errorf("String() = %v, want %v", got, want)
		}
	})

	t.Run("not empty", func(t *testing.T) {
		sourceValue := "My spoon is too big"

		pointer := &sourceValue
		want := sourceValue

		if got := String(pointer); got != want {
			t.Errorf("String() = %v, want %v", got, want)
		}
	})
}

func TestIsEmptyString(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var pointer *string

		if got := IsEmptyString(pointer); got != true {
			t.Errorf("IsEmptyString() = %v, want %v", got, true)
		}
	})

	t.Run("not empty", func(t *testing.T) {
		sourceValue := "My spoon is too big"

		pointer := &sourceValue

		if got := IsEmptyString(pointer); got != false {
			t.Errorf("IsEmptyString() = %v, want %v", got, false)
		}
	})
}

func TestInt64(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		var pointer *int64
		want := int64(0)

		if got := Int64(pointer); got != want {
			t.Errorf("Int64() = %v, want %v", got, want)
		}
	})

	t.Run("not empty", func(t *testing.T) {
		sourceValue := int64(64)

		pointer := &sourceValue
		want := sourceValue

		if got := Int64(pointer); got != want {
			t.Errorf("Int64() = %v, want %v", got, want)
		}
	})
}
