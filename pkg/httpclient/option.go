package httpclient

type Option func(p *Http)

func WithReqApiUrl(apiUrl string) Option {
	return func(s *Http) {
		s.ApiUrl = apiUrl
	}
}

func WithMethod(method string) Option {
	return func(s *Http) {
		s.Method = method
	}
}

func WithReqParams(reqParams interface{}) Option {
	return func(s *Http) {
		s.ReqParams = reqParams
	}
}

func WithHeaders(headers map[string]string) Option {
	return func(s *Http) {
		s.Headers = headers
	}
}

func WithBearerToken(bearerToken string) Option {
	return func(s *Http) {
		s.BearerToken = bearerToken
	}
}

func WithSetRetryCount(setRetryCount int) Option {
	return func(s *Http) {
		s.SetRetryCount = setRetryCount
	}
}

func WithShowLog(showLog bool) Option {
	return func(s *Http) {
		s.ShowLog = showLog
	}
}

func WithSecretID(secretID string) Option {
	return func(s *Http) {
		s.SecretID = secretID
	}
}

func WithSecretKey(secretKey string) Option {
	return func(s *Http) {
		s.SecretKey = secretKey
	}
}
