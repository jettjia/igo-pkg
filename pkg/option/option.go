package option

// option is a generic design for the option pattern
// avoid defining a lot of structs like this in your code
// in general t should be a struct
type Option[T any] func(t *T)

// apply applies opts on top of t
func Apply[T any](t *T, opts ...Option[T]) {
	for _, opt := range opts {
		opt(t)
	}
}

// option err behaves like option but returns an error
// you should use option in preference unless you need to do some validation when designing the option pattern
type OptionErr[T any] func(t *T) error

// apply err is like apply which applies opts on top of t
// if any of the opts returns an error then it breaks and returns an error
func ApplyErr[T any](t *T, opts ...OptionErr[T]) error {
	for _, opt := range opts {
		if err := opt(t); err != nil {
			return err
		}
	}
	return nil
}
