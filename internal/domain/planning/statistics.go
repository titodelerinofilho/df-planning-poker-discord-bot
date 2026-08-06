package planning

import (
	"fmt"
	"sort"
)

type NumericStatistics struct {
	HasNumericResult bool
	Min              int
	Max              int
	Median           float64
	Mode             []int
}

func CalculateNumericStatistics(scale Scale, estimates []Estimate) (NumericStatistics, error) {
	values := make([]int, 0, len(estimates))

	for _, estimate := range estimates {
		value, ok, err := scale.NumericValue(estimate)

		if err != nil {
			return NumericStatistics{}, fmt.Errorf("calculate numeric statistics: %w", err)
		}

		if !ok {
			continue
		}

		values = append(values, value)
	}

	if len(values) == 0 {
		return NumericStatistics{}, nil
	}

	sort.Ints(values)

	return NumericStatistics{
		HasNumericResult: true,
		Min:              values[0],
		Max:              values[len(values)-1],
		Median:           median(values),
		Mode:             mode(values),
	}, nil
}

func median(values []int) float64 {
	middle := len(values) / 2

	if len(values)%2 == 1 {
		return float64(values[middle])
	}

	return float64(values[middle-1]+values[middle]) / 2
}

func mode(values []int) []int {
	frequencies := make(map[int]int, len(values))
	highestFrequency := 0

	for _, value := range values {
		frequencies[value]++

		if frequencies[value] > highestFrequency {
			highestFrequency = frequencies[value]
		}
	}

	result := make([]int, 0)

	for value, frequency := range frequencies {
		if frequency == highestFrequency {
			result = append(result, value)
		}
	}

	sort.Ints(result)

	return result
}
