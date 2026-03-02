package gstr

// KebabCase converts a string into kebab case.
func KebabCase(s string) string {
	return delimiterCase(s, '-', false)
}

// UpperKebabCase converts a string into kebab case with capital letters.
func UpperKebabCase(s string) string {
	return delimiterCase(s, '-', true)
}
