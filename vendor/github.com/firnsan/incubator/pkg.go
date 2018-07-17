package incubator

type Application interface {
	GetVersion() string
	GetUsage() string

	OnReload() error
	OnOptParsed(map[string]interface{})

	OnStop()
}

func Incubate(app Application) {
	imp := newIncubator(app)
	imp.incubate()
}
