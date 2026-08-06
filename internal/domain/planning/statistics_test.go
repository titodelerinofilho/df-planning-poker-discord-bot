package planning

import (
	"errors"
	"reflect"
	"testing"
)

func TestCalculateNumericStatistics(t *testing.T) {
	scale := ModifiedFibonacciScale()
	tests := []struct {
		name      string
		values    []string
		wantStats NumericStatistics
	}{
		{
			name:   "single numeric value",
			values: []string{"5"},
			wantStats: NumericStatistics{
				HasNumericResult: true,
				Min:              5,
				Max:              5,
				Median:           5,
				Mode:             []int{5},
			},
		},
		{
			name:   "odd numeric values",
			values: []string{"1", "8", "3"},
			wantStats: NumericStatistics{
				HasNumericResult: true,
				Min:              1,
				Max:              8,
				Median:           3,
				Mode:             []int{1, 3, 8},
			},
		},
		{
			name:   "even numeric values",
			values: []string{"1", "2", "3", "5"},
			wantStats: NumericStatistics{
				HasNumericResult: true,
				Min:              1,
				Max:              5,
				Median:           2.5,
				Mode:             []int{1, 2, 3, 5},
			},
		},
		{
			name:   "mode with repeated value",
			values: []string{"1", "2", "2", "5", "5", "5", "8"},
			wantStats: NumericStatistics{
				HasNumericResult: true,
				Min:              1,
				Max:              8,
				Median:           5,
				Mode:             []int{5},
			},
		},
		{
			name:   "multiple modes",
			values: []string{"1", "1", "3", "3", "5"},
			wantStats: NumericStatistics{
				HasNumericResult: true,
				Min:              1,
				Max:              5,
				Median:           3,
				Mode:             []int{1, 3},
			},
		},
		{
			name:   "ignores special values",
			values: []string{"?", "8", "\u2615", "3"},
			wantStats: NumericStatistics{
				HasNumericResult: true,
				Min:              3,
				Max:              8,
				Median:           5.5,
				Mode:             []int{3, 8},
			},
		},
		{
			name:   "only special values",
			values: []string{"?", "\u2615"},
			wantStats: NumericStatistics{
				HasNumericResult: false,
			},
		},
		{
			name:   "no values",
			values: nil,
			wantStats: NumericStatistics{
				HasNumericResult: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats, err := CalculateNumericStatistics(scale, estimates(tt.values...))

			if err != nil {
				t.Fatalf("CalculateNumericStatistics() error = %v", err)
			}

			if !reflect.DeepEqual(stats, tt.wantStats) {
				t.Fatalf("CalculateNumericStatistics() = %#v, want %#v", stats, tt.wantStats)
			}
		})
	}
}

func TestCalculateNumericStatisticsReturnsInvalidEstimateError(t *testing.T) {
	scale := ModifiedFibonacciScale()

	_, err := CalculateNumericStatistics(scale, estimates("1", "4"))

	if !errors.Is(err, ErrInvalidEstimate) {
		t.Fatalf("CalculateNumericStatistics() error = %v, want ErrInvalidEstimate", err)
	}
}

func TestCalculateNumericStatisticsDoesNotMutateInput(t *testing.T) {
	scale := ModifiedFibonacciScale()
	input := estimates("8", "1", "3")
	before := []string{input[0].String(), input[1].String(), input[2].String()}

	_, err := CalculateNumericStatistics(scale, input)

	if err != nil {
		t.Fatalf("CalculateNumericStatistics() error = %v", err)
	}

	after := []string{input[0].String(), input[1].String(), input[2].String()}

	if !reflect.DeepEqual(after, before) {
		t.Fatalf("input = %v, want %v", after, before)
	}
}

func estimates(values ...string) []Estimate {
	result := make([]Estimate, 0, len(values))

	for _, value := range values {
		result = append(result, NewEstimate(value))
	}

	return result
}
