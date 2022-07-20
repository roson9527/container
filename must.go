package container

// MustSingleton wraps the `Singleton` method and panics on errors instead of returning the errors.
func MustSingleton(c Container, resolver interface{}) {
	must(c.Singleton(resolver))
}

// MustNamedSingleton wraps the `NamedSingleton` method and panics on errors instead of returning the errors.
func MustNamedSingleton(c Container, name string, resolver interface{}) {
	must(c.NamedSingleton(name, resolver))
}

// MustTransient wraps the `Transient` method and panics on errors instead of returning the errors.
func MustTransient(c Container, resolver interface{}) {
	must(c.Transient(resolver))
}

// MustNamedTransient wraps the `NamedTransient` method and panics on errors instead of returning the errors.
func MustNamedTransient(c Container, name string, resolver interface{}) {
	must(c.NamedTransient(name, resolver))
}

// MustCall wraps the `Call` method and panics on errors instead of returning the errors.
func MustCall(c Container, receiver interface{}) {
	must(c.Call(receiver))
}

// MustResolve wraps the `Resolve` method and panics on errors instead of returning the errors.
func MustResolve(c Container, abstraction interface{}) {
	must(c.Resolve(abstraction))
}

// MustNamedResolve wraps the `NamedResolve` method and panics on errors instead of returning the errors.
func MustNamedResolve(c Container, name string, abstraction interface{}) {
	must(c.NamedResolve(name, abstraction))
}

// MustFill wraps the `Fill` method and panics on errors instead of returning the errors.
func MustFill(c Container, receiver interface{}) {
	must(c.Fill(receiver))
}

func must(err error) {
	if err == nil {
		return
	}
	panic(err)
}
