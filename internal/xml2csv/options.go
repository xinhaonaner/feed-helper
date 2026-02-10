package xml2csv

// Option configures Convert behavior.
type Option func(*config)

type config struct {
	RowTag string // XML element name that represents one CSV row (e.g. "item", "record")
}

// RowTag sets the element name that defines one row. Default is "item".
func RowTag(tag string) Option {
	return func(c *config) {
		if tag != "" {
			c.RowTag = tag
		}
	}
}

func applyOptions(opts []Option) *config {
	c := &config{RowTag: "item"}
	for _, o := range opts {
		o(c)
	}
	return c
}
