package distconf

// A BackingLoader should run a single time and get a Reader for Config
type BackingLoader interface {
	Get() (Reader, error)
}

// BackingLoaderFunc can wrap a function to turn it into a BackingLoader
type BackingLoaderFunc func() (Reader, error)

// Get a Reader for Config, or an error if the Reader cannot be loaded
func (f BackingLoaderFunc) Get() (Reader, error) {
	return f()
}

// FromLoaders creates a Config from an array of loaders, only using loaders that don't load with
// error
func FromLoaders(loaders []BackingLoader, onFailure func(err error, loader BackingLoader)) *Distconf {
	readers := make([]Reader, 0, len(loaders))
	for _, l := range loaders {
		r, err := l.Get()
		if err != nil {
			if onFailure != nil {
				onFailure(err, l)
			}
			continue
		}
		readers = append(readers, r)
	}
	return New(readers)
}
