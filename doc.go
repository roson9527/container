package container

type BindType int

const (
	singletonType BindType = iota
	transientType
	delaySingletonType
)
