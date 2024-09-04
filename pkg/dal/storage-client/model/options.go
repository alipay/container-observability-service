package model

import "time"

type Options struct {
	OrderBy   string
	Ascending bool
	Limit     int
	From      time.Time
	To        time.Time
}
type OptionFunc func(*Options)

func WithLimit(limint int) OptionFunc {
	return func(o *Options) {
		o.Limit = limint
	}
}
func WithOrderBy(orderBy string) OptionFunc {
	return func(o *Options) {
		o.OrderBy = orderBy
	}
}
func WithDescending(descending bool) OptionFunc {
	return func(o *Options) {
		o.Ascending = descending
	}
}
func WithFrom(from time.Time) OptionFunc {
	return func(o *Options) {
		o.From = from
	}
}
func WithTo(to time.Time) OptionFunc {
	return func(o *Options) {
		o.To = to
	}
}
