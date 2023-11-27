package service

import "github.com/getsentry/sentry-go"

type options struct {
	temporalService   bool
	httpService       bool
	grpcService       bool
	traceSampleRate   float64
	profileSameplRate float64
	traceSampler      sentry.TracesSampler
	beforeSend        func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event
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

func WithTraceSampleRate(r float64) func(*options) {
	return func(o *options) {
		o.traceSampleRate = r
	}
}

func WithProfileSampleRate(r float64) func(*options) {
	return func(o *options) {
		o.profileSameplRate = r
	}
}

func WithTraceSampler(t sentry.TracesSampler) func(o *options) {
	return func(o *options) {
		o.traceSampler = t
	}
}

func WithBeforeSendTransaction(t func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event) func(o *options) {
	return func(o *options) {
		o.beforeSend = t
	}
}
