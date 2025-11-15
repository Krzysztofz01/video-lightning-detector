package utils

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	rangeExpressionSeparatorToken  string = ","
	rangeExpressionSeriesToken     string = "-"
	boundsExpressionSeparatorToken string = ":"
)

func IsRangeExpressionValid(expr string) bool {
	_, err := ParseRangeExpression(expr)
	return err == nil
}

func ParseRangeExpression(expr string) ([]int, error) {
	tokens := strings.Split(expr, rangeExpressionSeparatorToken)

	values := make([]int, 0)
	for _, token := range tokens {
		if strings.Contains(token, rangeExpressionSeriesToken) {
			if innerValues, err := parseRangeExpressionSeries(token); err != nil {
				return nil, fmt.Errorf("utils: failed to parse the range expression: %w", err)
			} else {
				values = append(values, innerValues...)
			}
		} else {
			if innerValue, err := parseRangeExpressionToken(token); err != nil {
				return nil, fmt.Errorf("utils: failed to parse the range expression: %w", err)
			} else {
				values = append(values, innerValue)
			}
		}
	}

	if isSorted := sort.SliceIsSorted(values, func(i, j int) bool { return i < j }); !isSorted {
		return nil, fmt.Errorf("utils: failed to parse the range expression due to invalid order")
	}

	return values, nil
}

func parseRangeExpressionToken(token string) (int, error) {
	if tokenValue, err := strconv.ParseInt(token, 10, 0); err != nil {
		return 0, fmt.Errorf("utils: failed to parse the range expression token: %w", err)
	} else {
		return int(tokenValue), nil
	}
}

func parseRangeExpressionSeries(token string) ([]int, error) {
	tokenParts := strings.Split(token, rangeExpressionSeriesToken)
	if len(tokenParts) != 2 {
		return nil, fmt.Errorf("utils: failed to parse token series value due to invlid format")
	}

	min, err := parseRangeExpressionToken(tokenParts[0])
	if err != nil {
		return nil, fmt.Errorf("utils: failed to parse series min token: %w", err)
	}

	max, err := parseRangeExpressionToken(tokenParts[1])
	if err != nil {
		return nil, fmt.Errorf("utils: failed to parse series max token: %w", err)
	}

	if min >= max {
		return nil, fmt.Errorf("utils: failed to parse token series value due to invlid value range")
	}

	values := make([]int, 0, max-min+1)
	for value := min; value <= max; value += 1 {
		values = append(values, value)
	}

	return values, nil
}

func IsBoundsExpressionValid(expr string) bool {
	_, _, _, _, err := ParseBoundsExpression(expr)
	return err == nil
}

func ParseBoundsExpression(expr string) (int, int, int, int, error) {
	tokens := strings.Split(expr, boundsExpressionSeparatorToken)
	if len(tokens) != 4 {
		return 0, 0, 0, 0, fmt.Errorf("utils: invalid bounds expression format")
	}

	x, err := parseBoundsExpressionToken(tokens[0], 0)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("utils: failed to parse the anchor point x value: %w", err)
	}

	y, err := parseBoundsExpressionToken(tokens[1], 0)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("utils: failed to parse the anchor point y value: %w", err)
	}

	w, err := parseBoundsExpressionToken(tokens[2], 1)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("utils: failed to parse the dimension w value: %w", err)
	}

	h, err := parseBoundsExpressionToken(tokens[3], 1)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("utils: failed to parse the dimension h value: %w", err)
	}

	return x, y, w, h, nil
}

func parseBoundsExpressionToken(token string, min int) (int, error) {
	value64, err := strconv.ParseInt(token, 10, 0)
	if err != nil {
		return 0, fmt.Errorf("utils: failed to parse bounds expression token: %w", err)
	}

	value := int(value64)
	if value < min {
		return 0, fmt.Errorf("utils: the bound expression token value is out of range: %w", err)
	}

	return value, nil
}
