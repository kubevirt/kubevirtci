package opts

type Opt interface {
	Exec() error
}
