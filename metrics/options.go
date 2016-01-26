package metrics

type Options struct {
	Namespace string
}

func Namespace(n string) Option {
	return func(o *Options) {
		o.Namespace = n
	}
}
