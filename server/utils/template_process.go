package server_utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

func ProcessTemplateJSON(template interface{}, ctx EContext) (interface{}, error) {
	switch t := template.(type) {

	case string:
		trimmed := strings.TrimSpace(t)
		re := regexp.MustCompile(`{{\s*([a-zA-Z0-9_.-]+)([^}]*)}}`)

		// state.xxx shortcut handling
		if matches := re.FindStringSubmatch(trimmed); len(matches) > 1 && trimmed == matches[0] && ctx.State != nil {
			switch matches[1] {
			case "state.list":
				return ctx.State.List, nil
			case "state.item":
				return ctx.State.Item, nil
			case "state.created":
				return ctx.State.Created, nil
			case "state.updated":
				return ctx.State.Updated, nil
			}
		}

		// Normal template replacement
		result := re.ReplaceAllStringFunc(t, func(match string) string {
			parts := re.FindStringSubmatch(match)
			if len(parts) < 2 {
				return match
			}
			key := parts[1]
			args := strings.TrimSpace(parts[2])

			// request values
			if strings.HasPrefix(key, "request.") {
				val, err := evalResolveValue(key, ctx)
				if err == nil {
					return fmt.Sprintf("%v", val)
				}
				return match
			}

			// Faker process
			switch key {
			case "name":
				return gofakeit.Name()
			case "uuid":
				return gofakeit.UUID()
			case "email":
				return gofakeit.Email()
			case "bool":
				return fmt.Sprintf("%v", gofakeit.Bool())
			case "date":
				return gofakeit.Date().Format("2006-01-02")
			case "dateFuture":
				days := 1
				fmt.Sscanf(args, "days=%d", &days)
				return gofakeit.DateRange(time.Now(), time.Now().AddDate(0, 0, days)).Format("2006-01-02")
			case "dateNow":
				return gofakeit.DateRange(time.Now(), time.Now().AddDate(0, 0, 0)).Format("2006-01-02")
			case "number":
				min, max := 1, 1000
				fmt.Sscanf(args, "min=%d max=%d", &min, &max)
				return fmt.Sprintf("%d", gofakeit.Number(min, max))
			default:
				return match
			}
		})

		return result, nil

	case map[string]interface{}:
		res := make(map[string]interface{}, len(t))
		for k, v := range t {
			processed, err := ProcessTemplateJSON(v, ctx)
			if err != nil {
				return nil, err
			}
			res[k] = processed
		}
		return res, nil

	case []interface{}:
		res := make([]interface{}, len(t))
		for i, v := range t {
			processed, err := ProcessTemplateJSON(v, ctx)
			if err != nil {
				return nil, err
			}
			res[i] = processed
		}
		return res, nil

	default:
		return t, nil
	}
}
