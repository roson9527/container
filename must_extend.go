package container

func MustResolves(c Container, abstraction ...any) {
	for _, a := range abstraction {
		must(c.Resolve(a))
	}
}
