package planning

import (
	"errors"
	"testing"
)

func TestModifiedFibonacciScaleValues(t *testing.T) {
	scale := ModifiedFibonacciScale()
	values := scale.Values()
	want := []string{"0", "1", "2", "3", "5", "8", "13", "21", "34", "55", "?", "\u2615"}

	if len(values) != len(want) {
		t.Fatalf("values = %d, want %d", len(values), len(want))
	}

	for index, estimate := range values {
		if estimate.String() != want[index] {
			t.Fatalf("values[%d] = %q, want %q", index, estimate.String(), want[index])
		}
	}
}

func TestModifiedFibonacciScaleValidate(t *testing.T) {
	scale := ModifiedFibonacciScale()
	tests := []struct {
		name  string
		value string
		valid bool
	}{
		{name: "zero", value: "0", valid: true},
		{name: "one", value: "1", valid: true},
		{name: "two", value: "2", valid: true},
		{name: "three", value: "3", valid: true},
		{name: "five", value: "5", valid: true},
		{name: "eight", value: "8", valid: true},
		{name: "thirteen", value: "13", valid: true},
		{name: "twenty one", value: "21", valid: true},
		{name: "thirty four", value: "34", valid: true},
		{name: "fifty five", value: "55", valid: true},
		{name: "unknown", value: "?", valid: true},
		{name: "coffee", value: "\u2615", valid: true},
		{name: "empty", value: "", valid: false},
		{name: "four", value: "4", valid: false},
		{name: "negative", value: "-1", valid: false},
		{name: "word", value: "small", valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := scale.Validate(NewEstimate(tt.value))

			if tt.valid && err != nil {
				t.Fatalf("Validate() error = %v, want nil", err)
			}

			if !tt.valid && !errors.Is(err, ErrInvalidEstimate) {
				t.Fatalf("Validate() error = %v, want ErrInvalidEstimate", err)
			}
		})
	}
}

func TestModifiedFibonacciScaleOrdinal(t *testing.T) {
	scale := ModifiedFibonacciScale()
	tests := []struct {
		value   string
		ordinal int
	}{
		{value: "0", ordinal: 0},
		{value: "1", ordinal: 1},
		{value: "2", ordinal: 2},
		{value: "3", ordinal: 3},
		{value: "5", ordinal: 4},
		{value: "8", ordinal: 5},
		{value: "13", ordinal: 6},
		{value: "21", ordinal: 7},
		{value: "34", ordinal: 8},
		{value: "55", ordinal: 9},
		{value: "?", ordinal: 10},
		{value: "\u2615", ordinal: 11},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			ordinal, err := scale.Ordinal(NewEstimate(tt.value))

			if err != nil {
				t.Fatalf("Ordinal() error = %v", err)
			}

			if ordinal != tt.ordinal {
				t.Fatalf("Ordinal() = %d, want %d", ordinal, tt.ordinal)
			}
		})
	}
}

func TestModifiedFibonacciScaleSpecialValues(t *testing.T) {
	tests := []struct {
		value   string
		special bool
	}{
		{value: "0", special: false},
		{value: "55", special: false},
		{value: "?", special: true},
		{value: "\u2615", special: true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			estimate := NewEstimate(tt.value)

			if estimate.IsSpecial() != tt.special {
				t.Fatalf("IsSpecial() = %v, want %v", estimate.IsSpecial(), tt.special)
			}
		})
	}
}

func TestModifiedFibonacciScaleNumericValue(t *testing.T) {
	scale := ModifiedFibonacciScale()
	tests := []struct {
		value   string
		numeric int
		ok      bool
	}{
		{value: "0", numeric: 0, ok: true},
		{value: "1", numeric: 1, ok: true},
		{value: "55", numeric: 55, ok: true},
		{value: "?", ok: false},
		{value: "\u2615", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			numeric, ok, err := scale.NumericValue(NewEstimate(tt.value))

			if err != nil {
				t.Fatalf("NumericValue() error = %v", err)
			}

			if ok != tt.ok {
				t.Fatalf("NumericValue() ok = %v, want %v", ok, tt.ok)
			}

			if numeric != tt.numeric {
				t.Fatalf("NumericValue() numeric = %d, want %d", numeric, tt.numeric)
			}
		})
	}
}

func TestModifiedFibonacciScaleRejectsInvalidOrdinalAndNumericValue(t *testing.T) {
	scale := ModifiedFibonacciScale()
	estimate := NewEstimate("4")

	_, err := scale.Ordinal(estimate)

	if !errors.Is(err, ErrInvalidEstimate) {
		t.Fatalf("Ordinal() error = %v, want ErrInvalidEstimate", err)
	}

	_, _, err = scale.NumericValue(estimate)

	if !errors.Is(err, ErrInvalidEstimate) {
		t.Fatalf("NumericValue() error = %v, want ErrInvalidEstimate", err)
	}
}

func TestNewEstimateTrimsValue(t *testing.T) {
	estimate := NewEstimate("  13  ")

	if estimate.String() != "13" {
		t.Fatalf("String() = %q, want 13", estimate.String())
	}
}
