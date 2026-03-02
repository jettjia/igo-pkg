package xerror

type Option func(p *MyError)

func WithCause(cause string) Option {
	return func(s *MyError) {
		s.Cause = cause
	}
}

func WithSolution(solution string) Option {
	return func(s *MyError) {
		s.Solution = solution
	}
}

func WithDetail(detail interface{}) Option {
	return func(s *MyError) {
		s.Detail = detail
	}
}
