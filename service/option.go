package service

type options struct {
	temporalService bool
	httpService     bool
	grpcService     bool
}

// WithGRPCService will enable to service to run with a temporal worker
func WithGRPCService() func(*options) {
	return func(o *options) {
		o.grpcService = true
	}
}

// WithTemporalService will enable to service to run with a temporal worker
func WithTemporalService() func(*options) {
	return func(o *options) {
		o.temporalService = true
	}
}

// WithHTTPService will enable to service to run with an http service
func WithHTTPService() func(*options) {
	return func(o *options) {
		o.httpService = true
	}
}
