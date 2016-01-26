package metrics

type Options struct {
	Namespace string
	Fields    Fields
}

func Namespace(n string) Option {
	return func(o *Options) {
		o.Namespace = n
	}
}

func WithFields(f Fields) Option {
	return func(o *Options) {
		o.Fields = f
	}
}
