package planning

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	SpecialEstimateUnknown = "?"
	SpecialEstimateCoffee  = "\u2615"
)

var ErrInvalidEstimate = errors.New("invalid estimate")

type Estimate struct {
	value string
}

type Scale struct {
	values   []Estimate
	ordinals map[string]int
	numeric  map[string]int
}

func ModifiedFibonacciScale() Scale {
	values := []string{"0", "1", "2", "3", "5", "8", "13", "21", "34", "55", SpecialEstimateUnknown, SpecialEstimateCoffee}
	estimates := make([]Estimate, 0, len(values))
	ordinals := make(map[string]int, len(values))
	numeric := make(map[string]int, len(values))

	for index, value := range values {
		estimate := Estimate{value: value}
		estimates = append(estimates, estimate)
		ordinals[value] = index

		parsed, err := strconv.Atoi(value)

		if err == nil {
			numeric[value] = parsed
		}
	}

	return Scale{
		values:   estimates,
		ordinals: ordinals,
		numeric:  numeric,
	}
}

func NewEstimate(value string) Estimate {
	return Estimate{
		value: strings.TrimSpace(value),
	}
}

func (estimate Estimate) String() string {
	return estimate.value
}

func (estimate Estimate) IsSpecial() bool {
	return estimate.value == SpecialEstimateUnknown || estimate.value == SpecialEstimateCoffee
}

func (scale Scale) Values() []Estimate {
	values := make([]Estimate, len(scale.values))
	copy(values, scale.values)

	return values
}

func (scale Scale) Validate(estimate Estimate) error {
	if _, ok := scale.ordinals[estimate.value]; !ok {
		return fmt.Errorf("%w: %s", ErrInvalidEstimate, estimate.value)
	}

	return nil
}

func (scale Scale) Ordinal(estimate Estimate) (int, error) {
	ordinal, ok := scale.ordinals[estimate.value]

	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrInvalidEstimate, estimate.value)
	}

	return ordinal, nil
}

func (scale Scale) NumericValue(estimate Estimate) (int, bool, error) {
	err := scale.Validate(estimate)

	if err != nil {
		return 0, false, err
	}

	value, ok := scale.numeric[estimate.value]

	return value, ok, nil
}
