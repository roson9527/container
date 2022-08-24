// Package container is a lightweight yet powerful IoC container for Go projects.
// It provides an easy-to-use interface and performance-in-mind container to be your ultimate requirement.
package container

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

// binding holds a resolver and a concrete (if singleton).
// It is the break for the Container wall!
type binding struct {
	bindType BindType
	resolver interface{} // resolver is the function that is responsible for making the concrete.
	concrete interface{} // concrete is the stored instance for singleton bindings.
}

// make resolves the binding if needed and returns the resolved concrete.
func (b *binding) make(c Container, opt *Option) (interface{}, error) {
	if b.concrete != nil {
		return b.concrete, nil
	}

	if b.bindType == delaySingletonType {
		var err error
		b.concrete, err = c.invoke(b.resolver, opt)
		if err != nil {
			return nil, err
		}
		return b.concrete, nil
	}

	return c.invoke(b.resolver, opt)
}

// Container holds the bindings and provides methods to interact with them.
// It is the entry point in the package.
// Use a pointer to make it lazily change state
type Container map[reflect.Type]map[string]*binding

// New creates a new concrete of the Container.
func New() Container {
	return make(Container)
}

// bind maps an abstraction to concrete and instantiates if it is a singleton binding.
func (c Container) bind(resolver interface{}, bindType BindType, opt *Option) error {
	reflectedResolver := reflect.TypeOf(resolver)
	if reflectedResolver.Kind() != reflect.Func {
		return errors.New("container: the resolver must be a function")
	}

	// 输出参数简化为只返回一个，如果有必要可以修改为多个
	if reflectedResolver.NumOut() > 0 {
		if _, exist := c[reflectedResolver.Out(0)]; !exist {
			c[reflectedResolver.Out(0)] = make(map[string]*binding)
		}
	}

	var concrete interface{}
	var err error
	if bindType == singletonType {
		concrete, err = c.invoke(resolver, opt)
		if err != nil {
			return err
		}
	} else {
		concrete = nil
	}
	if c[reflectedResolver.Out(0)][opt.name] != nil {
		rType := reflectedResolver.Out(0)
		name := opt.name
		if opt.name == "" {
			name = "type"
		}
		return fmt.Errorf("container: %s binding [%s] already exists", rType.String(), name)
	}

	c[reflectedResolver.Out(0)][opt.name] = &binding{resolver: resolver, concrete: concrete, bindType: bindType}

	return nil
}

// invoke calls a function and its returned values.
// It only accepts one value and an optional error.
func (c Container) invoke(function interface{}, opt *Option) (interface{}, error) {
	arguments, err := c.arguments(function, opt)
	if err != nil {
		return nil, err
	}

	values := reflect.ValueOf(function).Call(arguments)

	if len(values) == 1 || len(values) == 2 {
		if len(values) == 2 && values[1].CanInterface() {
			if err, ok := values[1].Interface().(error); ok {
				return values[0].Interface(), err
			}
		}
		return values[0].Interface(), nil
	}

	return nil, errors.New("container: resolver function signature is invalid")
}

// arguments returns the list of resolved arguments for a function.
func (c Container) arguments(function interface{}, opt *Option) ([]reflect.Value, error) {
	reflectedFunction := reflect.TypeOf(function)
	argumentsCount := reflectedFunction.NumIn()
	arguments := make([]reflect.Value, argumentsCount)

	for i := 0; i < argumentsCount; i++ {
		abstraction := reflectedFunction.In(i)

		if concrete, exist := c.getBinding(abstraction, toNames(opt.name)); exist {
			concreteInstance, err := concrete.make(c, opt)
			if err != nil {
				return nil, err
			}
			arguments[i] = reflect.ValueOf(concreteInstance)
		} else {
			return nil, fmt.Errorf("container: no binding found for %s", abstraction.String())
		}
	}

	return arguments, nil
}

func (c Container) getBinding(t reflect.Type, names []string) (*binding, bool) {
	src := c[t]
	if c[t] == nil {
		panic(fmt.Sprintf("container: no binding found for %s", t.String()))
	}

	for i := 0; i < len(names); i++ {
		if val, ok := src[names[i]]; ok {
			return val, true
		}
	}

	return nil, false
}

// Reset deletes all the existing bindings and empties the container.
func (c Container) Reset() {
	for k := range c {
		delete(c, k)
	}
}

func (c Container) singleton(resolver interface{}, opt *Option) error {
	if opt.delay {
		return c.bind(resolver, delaySingletonType, opt)
	} else {
		return c.bind(resolver, singletonType, opt)
	}
}

// Singleton binds an abstraction to concrete in singleton mode.
// It takes a resolver function that returns the concrete, and its return type matches the abstraction (interface).
// The resolver function can have arguments of abstraction that have been declared in the Container already.
func (c Container) Singleton(resolver interface{}, opt ...OptionFunc) error {
	option := LoadOption(opt...)
	return c.singleton(resolver, option)
}

// NamedSingleton binds a named abstraction to concrete in singleton mode.
func (c Container) NamedSingleton(name string, resolver interface{}, opt ...OptionFunc) error {
	option := LoadOption(opt...)
	option.name = name
	return c.singleton(resolver, option)
}

func (c Container) transient(resolver interface{}, opt *Option) error {
	return c.bind(resolver, transientType, opt)
}

// Transient binds an abstraction to concrete in transient mode.
// It takes a resolver function that returns the concrete, and its return type matches the abstraction (interface).
// The resolver function can have arguments of abstraction that have been declared in the Container already.
func (c Container) Transient(resolver interface{}, opt ...OptionFunc) error {
	option := LoadOption(opt...)
	return c.transient(resolver, option)
}

// NamedTransient binds a named abstraction to concrete in transient mode.
func (c Container) NamedTransient(name string, resolver interface{}, opt ...OptionFunc) error {
	option := LoadOption(opt...)
	option.name = name
	return c.transient(resolver, option)
}

// Call takes a receiver function with one or more arguments of the abstractions (interfaces).
// It invokes the receiver function and passes the related concretes.
func (c Container) Call(function interface{}, opt ...OptionFunc) error {
	receiverType := reflect.TypeOf(function)
	if receiverType == nil || receiverType.Kind() != reflect.Func {
		return errors.New("container: invalid function")
	}

	option := LoadOption(opt...)
	arguments, err := c.arguments(function, option)
	if err != nil {
		return err
	}

	result := reflect.ValueOf(function).Call(arguments)

	if len(result) == 0 {
		return nil
	} else if len(result) == 1 && result[0].CanInterface() {
		if result[0].IsNil() {
			return nil
		}
		if err, ok := result[0].Interface().(error); ok {
			return err
		}
	}

	return errors.New("container: receiver function signature is invalid")
}

// 填充指针
func (c Container) resolve(abstraction interface{}, opt *Option) error {
	receiverType := reflect.TypeOf(abstraction)
	if receiverType == nil {
		return errors.New("container: invalid abstraction")
	}

	if receiverType.Kind() == reflect.Ptr {
		elem := receiverType.Elem()

		if concrete, exist := c.getBinding(elem, toNames(opt.name)); exist {
			if instance, err := concrete.make(c, opt); err == nil {
				reflect.ValueOf(abstraction).Elem().Set(reflect.ValueOf(instance))
				return nil
			} else {
				return err
			}
		}

		return errors.New("container: no concrete found for: " + elem.String())
	}

	return errors.New("container: invalid abstraction")
}

// Resolve takes an abstraction (reference of an interface type) and fills it with the related concrete.
func (c Container) Resolve(abstraction interface{}, opt ...OptionFunc) error {
	option := LoadOption(opt...)
	return c.resolve(abstraction, option)
}

// NamedResolve takes abstraction and its name and fills it with the related concrete.
func (c Container) NamedResolve(name string, abstraction interface{}, opt ...OptionFunc) error {
	option := LoadOption(opt...)
	option.name = name
	return c.resolve(abstraction, option)
}

// Fill takes a struct and resolves the fields with the tag `container:"inject"`
func (c Container) Fill(structure interface{}) error {
	receiverType := reflect.TypeOf(structure)
	if receiverType == nil {
		return errors.New("container: invalid structure")
	}

	opt := defaultOption()

	if receiverType.Kind() == reflect.Ptr {
		elem := receiverType.Elem()
		if elem.Kind() == reflect.Struct {
			s := reflect.ValueOf(structure).Elem()

			for i := 0; i < s.NumField(); i++ {
				f := s.Field(i)

				// 使用了新的匹配方式
				// container:type -> 按类型进行匹配
				// container:name -> 按类型+名称进行匹配（外部可访问的属性名字）
				if t, exist := s.Type().Field(i).Tag.Lookup("container"); exist {
					subTs := strings.Split(t, ",")
					names := make([]string, 0)

					if len(subTs) == 0 {
						names = append(names, "")
					} else {
						for _, subT := range subTs {
							switch subT {
							case "type":
								names = append(names, "")
							case "name":
								names = append(names, s.Type().Field(i).Name)
							default:
								names = append(names, subT)
							}
						}
					}

					if concrete, exist := c.getBinding(f.Type(), names); exist {
						instance, _ := concrete.make(c, opt)

						ptr := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
						ptr.Set(reflect.ValueOf(instance))

						continue
					}

					return errors.New(fmt.Sprintf("container: cannot make %v(%v) field with tags [%s]",
						s.Type().Field(i).Name, f.Type().String(), strings.Join(names, ",")))
				}
			}

			return nil
		}
	}

	return errors.New("container: invalid structure")
}
