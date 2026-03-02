package builder

import "strings"

func BuildSelect(selectField []string) (string, error) {
	fields := "*"
	if len(selectField) == 0 {
		return fields, nil
	}

	fields = strings.Join(selectField, ",")
	bd := strings.Builder{}
	bd.WriteString(fields)

	return bd.String(), nil
}

func BuildSelectVariable(selectArgs ...[]string) (selectField []string) {
	if len(selectArgs) == 0 {
		selectField = []string{"*"}

		return
	}
	for _, arg := range selectArgs {
		selectField = append(selectField, arg...)
	}

	return
}
