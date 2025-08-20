package server_utils

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// FilteredMockData applies filtering, sorting, and pagination
// to a JSON-like slice of objects.
//
// Processing order:
//  1. Exact filters   (?field=value)
//  2. "Like" filters  (?field_like=value)
//  3. Sorting         (?_sort=field&_order=asc|desc)
//  4. Pagination      (?_page=n&_limit=m)
//
// Returns the transformed slice or an error if pagination parameters are invalid.
func FilteredMockData(data []map[string]interface{}, params map[string]string) ([]map[string]interface{}, error) {
	filtered := data

	filtered = applyExactFilters(filtered, params)

	filtered = applyLikeFilters(filtered, params)

	applySorting(filtered, params)

	filtered, err := applyPagination(filtered, params)
	if err != nil {
		return nil, err
	}

	return filtered, nil
}

// Slices the dataset into pages using
// query parameters `_page` and `_limit`.
// Returns an error if parameters are invalid.
func applyPagination(data []map[string]interface{}, params map[string]string) ([]map[string]interface{}, error) {
	limit := 0
	page := 1
	if val, ok := params["_limit"]; ok {
		if _, err := fmt.Sscanf(val, "%d", &limit); err != nil || limit < 0 {
			return nil, fmt.Errorf("_limit must be a positive number")
		}
	}
	if val, ok := params["_page"]; ok {
		if _, err := fmt.Sscanf(val, "%d", &page); err != nil || page < 1 {
			return nil, fmt.Errorf("_page must be a positive number")
		}
	}

	// No pagination requested
	if limit <= 0 {
		return data, nil
	}

	start := (page - 1) * limit
	if start >= len(data) {
		return []map[string]interface{}{}, nil
	}

	end := start + limit
	if end > len(data) {
		end = len(data)
	}
	return data[start:end], nil
}

// matchExact checks strict equality between a given value and a target string.
// Supports float64, string, and bool comparisons.
func matchExact(v interface{}, target string) bool {
	switch val := v.(type) {
	case float64:
		return target == fmt.Sprintf("%.0f", val)
	case string:
		return val == target
	case bool:
		return (target == "true" && val) || (target == "false" && !val)
	default:
		return false
	}
}

func applyExactFilters(data []map[string]interface{}, params map[string]string) []map[string]interface{} {
	filtered := data
	for key, val := range params {
		if strings.HasPrefix(key, "_") || key == "apiKey" || strings.HasSuffix(key, "_like") {
			continue
		}

		decodedVal, _ := url.QueryUnescape(val)
		tmp := []map[string]interface{}{}

		for _, item := range filtered {
			if v, ok := item[key]; ok {
				if matchExact(v, decodedVal) {
					tmp = append(tmp, item)
				}
			}
		}
		filtered = tmp
	}
	return filtered
}

func applyLikeFilters(data []map[string]interface{}, params map[string]string) []map[string]interface{} {
	filtered := data
	for key, val := range params {
		if !strings.HasSuffix(key, "_like") {
			continue
		}

		field := strings.TrimSuffix(key, "_like")
		decodedVal, _ := url.QueryUnescape(val)
		needle := strings.ToLower(decodedVal)

		tmp := []map[string]interface{}{}
		for _, item := range filtered {
			if v, ok := item[field]; ok {
				strVal := strings.ToLower(fmt.Sprintf("%v", v))
				if strings.Contains(strVal, needle) {
					tmp = append(tmp, item)
				}
			}
		}
		filtered = tmp
	}
	return filtered
}

func applySorting(data []map[string]interface{}, params map[string]string) {
	sortField := params["_sort"]
	sortOrder := strings.ToLower(params["_order"])
	if sortField == "" {
		return
	}

	sort.SliceStable(data, func(i, j int) bool {
		vi, ok1 := data[i][sortField]
		vj, ok2 := data[j][sortField]
		if !ok1 && !ok2 {
			return false
		}
		if !ok1 {
			return false
		}
		if !ok2 {
			return true
		}
		return compareValues(vi, vj, sortOrder)
	})
}

func compareValues(a, b interface{}, order string) bool {
	switch va := a.(type) {
	case float64:
		vb, _ := b.(float64)
		if order == "desc" {
			return va > vb
		}
		return va < vb
	case string:
		vb := fmt.Sprintf("%v", b)
		if order == "desc" {
			return va > vb
		}
		return va < vb
	case bool:
		vb, _ := b.(bool)
		if order == "desc" {
			return va && !vb
		}
		return !va && vb
	default:
		return true
	}
}
