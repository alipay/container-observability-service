package xsearch

type ElasticSearchConf struct {
	Endpoint string
	User     string
	Password string
	Index    string
	Type     string
}

var (
	XSearchClear Cleaner
)

type Cleaner []func()

func (c *Cleaner) AddCleanWork(w func()) {
	*c = append(*c, w)
}
func (c *Cleaner) DoClear() {
	for _, w := range *c {
		w()
	}
}
