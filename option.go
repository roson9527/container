package container

import "strings"

type Option struct {
	name  string
	delay bool
}

type OptionFunc func(*Option) error

func defaultOption() *Option {
	return &Option{
		name:  "",   // 空的时候为类型，否则为名称
		delay: true, // 默认延迟创建实例，避免依赖顺序而影响初始化
	}
}

func LoadOption(opt ...OptionFunc) *Option {
	option := defaultOption()
	for _, f := range opt {
		if err := f(option); err != nil {
			panic(err)
		}
	}
	return option
}

//
//func SetName(name string) OptionFunc {
//	return func(option *Option) error {
//		option.name = name
//		return nil
//	}
//}
//
//func SetDelay(delay bool) OptionFunc {
//	return func(option *Option) error {
//		option.delay = delay
//		return nil
//	}
//}

func toNames(src string) []string {
	// 如果为空那么则是降级寻找默认注册类型
	if len(src) == 0 {
		return []string{""}
	}

	tmp := strings.Split(strings.TrimSpace(src), ",")
	out := make([]string, 0, 0)
	for i := 0; i < len(tmp); i++ {
		t := strings.TrimSpace(tmp[i])
		if len(t) == 0 {
			continue
		}
		switch t {
		case "type":
			out = append(out, "")
		default:
			out = append(out, t)
		}
	}

	return out
}
