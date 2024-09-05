package model

import "time"

type SloOptions struct {
	Result         string
	Cluster        string
	BizName        string
	Count          string
	Type           string    // create 或者 delete, 默认是 create
	DeliveryStatus string    // FAIL/KILL/ALL/SUCCESS
	SloTime        string    // 20s/30m0s/10m0s
	Env            string    // prod, test
	From           time.Time // range query
	To             time.Time // range query
}
type NodeParams struct {
	NodeUid  string
	NodeIp   string
	NodeName string
}
type PodParams struct {
	Name     string
	Uid      string
	Hostname string
	Podip    string
	From     time.Time
	To       time.Time
}

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
