package conf

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
)

// 配置项不存在错误
type NotFoundErr struct {
	key string
}

func (n NotFoundErr) Error() string {
	return fmt.Sprintf("prop whith key: %s not fond", n.key)
}

// 是否为配置项不存在的错误
func NotFound(err error) bool {
	_, ok := err.(NotFoundErr)
	return ok
}

// 配置
type Config struct {
	sources []Source
}

// 添加一个配置源，最高优先级
func (c *Config) AddFirst(source Source) {
	// 扩容
	c.sources = append(c.sources, source)
	// 右移
	copy(c.sources[1:], c.sources[0:])
	// 设置
	c.sources[0] = source
}

// 添加一个配置源，最低优先级
func (c *Config) AddLast(source Source) {
	c.sources = append(c.sources, source)
}

// AddCommandLineSource 添加命令行的配置
func (c *Config) AddCommandLineSource() {
	c.AddLast(CMDLineSource())
}

// AddNacosSource 添加nacos配置
func (c *Config) AddNacosSource(nacosUrl, namespaceId, dataId, group, username, password string) error {
	source, err := NacosSource(nacosUrl, namespaceId, dataId, group, username, password)
	if err != nil {
		return err
	}
	c.AddLast(source)
	return nil
}

// AddNacosSourceFromConfig 从配置中获取参数添加nacos的配置
func (c *Config) AddNacosSourceFromConfig() error {
	dataId := c.GetStringDefault("nacos.dataId", "")
	namespace := c.GetStringDefault("nacos.namespaceId", "")
	nacosUrl := c.GetStringDefault("nacos.url", "")
	group := c.GetStringDefault("nacos.group", "DEFAULT_GROUP")
	username := c.GetStringDefault("nacos.username", "")
	password := c.GetStringDefault("nacos.password", "")

	if nacosUrl == "" || namespace == "" || dataId == "" {
		fmt.Println("did not load nacos config source,  because not found nacos params from config")
		return nil
	}

	return c.AddNacosSource(nacosUrl, namespace, dataId, group, username, password)
}

// AddFileSource 添加文件配置
func (c *Config) AddFileSource(file string) error {
	source, err := FileSource(file)

	if err != nil {
		return err
	}

	c.AddLast(source)
	return nil
}

// AddFileSourceFromConfig 从配置中获取参数添加文件配置
func (c *Config) AddFileSourceFromConfig() error {
	file := c.GetStringDefault("config.file", "")
	if file == "" {
		fmt.Println("did not load file config source, because not found config.file from config")
		return nil
	}
	return c.AddFileSource(file)
}

// 获取配置项的值, 不存在配置项则返回零值和NotFoundErr, error只可能是nil或NotFoundErr
func (c *Config) GetString(key string) (string, error) {
	for _, s := range c.sources {
		v := s.Get(key)
		if v != "" {
			return v, nil
		}
	}
	return "", NotFoundErr{key: key}
}

// 获取配置项的值, 不存在配置项则返回第二个参数
func (c *Config) GetStringDefault(key string, val string) string {
	v, err := c.GetString(key)
	// 如果有错误，这个错误只可能是NotFoundErr
	// 在这里如果没有找到的话就返回默认值
	if err != nil {
		return val
	}
	return v
}

// 获取配置项的值, 与GetString一样，除了遇到错误会panic
func (c *Config) MustGetString(key string) string {
	v, err := c.GetString(key)
	if err != nil {
		panic(fmt.Sprintf("config err: %v not found\n%+v", key, err))
	}
	return v
}

// 获取bool类型的配置项, 不存在配置项则返回零值和NotFoundErr
func (c *Config) GetBool(key string) (bool, error) {
	s, err := c.GetString(key)
	if err != nil {
		return false, err
	}
	v, err := strconv.ParseBool(s)
	if err != nil {
		return false, errors.Wrapf(err, "prop value: %s with key: %s is not bool value", s, key)
	}
	return v, nil
}

// 获取bool类型的配置项, 不存在配置项则返回第二个参数并且error==nil
func (c *Config) GetBoolDefault(key string, val bool) (bool, error) {
	b, err := c.GetBool(key)
	if err != nil && NotFound(err) {
		return val, nil
	}
	return b, err
}

// 获取bool类型的配置项，与GetBool一样，除了遇到错误会panic
func (c *Config) MustGetBool(key string) bool {
	v, err := c.GetBool(key)
	if err != nil {
		panic(fmt.Sprintf("config err: %v\n%+v", key, err))
	}
	return v
}

// 获取int类型的配置值, 不存在配置项则返回零值和NotFoundErr
func (c *Config) GetInt(key string) (int, error) {
	i, err := c.GetInt64(key)
	return int(i), err
}

// 获取int类型的配置项，不存在配置项则返回第二个参数并且error==nil
func (c *Config) GetIntDefault(key string, val int) (int, error) {
	i, err := c.GetInt(key)
	if err != nil && NotFound(err) {
		return val, nil
	}
	return i, err
}

// 回去int类型的配置项，与GetInt一样，除了遇到错误会panic
func (c *Config) MustGetInt(key string) int {
	i, err := c.GetInt(key)
	if err != nil {
		panic(fmt.Sprintf("config err: %v\n%+v", key, err))
	}
	return i
}

// 获取int64类型的配置项,不存在配置项则返回零值和NotFoundErr
func (c *Config) GetInt64(key string) (int64, error) {
	s, err := c.GetString(key)
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "prop value: %s with key: %s is not int value", s, key)
	}
	return v, nil
}

// 获取int64类型的配置项,不存在配置项则返回第二个参数并且error==nil
func (c *Config) GetInt64Default(key string, val int64) (int64, error) {
	i, err := c.GetInt64(key)
	if err != nil && NotFound(err) {
		return val, nil
	}
	return i, err
}

// 回去int64类型的配置项，与GetInt64一样，除了遇到错误会panic
func (c *Config) MustGetInt64(key string) int64 {
	i, err := c.GetInt64(key)
	if err != nil {
		panic(fmt.Sprintf("config err: %v\n%+v", key, err))
	}
	return i
}

// 获取uint64类型的配置项,不存在配置项则返回零值和NotFoundErr
func (c *Config) GetUint64(key string) (uint64, error) {
	s, err := c.GetString(key)
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "prop value: %s with key: %s is not uint value", s, key)
	}
	return v, nil
}

// 获取uint64类型的配置项,不存在配置项则返回第二个参数并且error==nil
func (c *Config) GetUint64Default(key string, val uint64) (uint64, error) {
	u, err := c.GetUint64(key)
	if err != nil && NotFound(err) {
		return val, err
	}
	return u, err
}

// 获取uint64类型的配置项，与GetUint64一样，除了遇到错误会panic
func (c *Config) MustGetUint64(key string) uint64 {
	u, err := c.GetUint64(key)
	if err != nil {
		panic(fmt.Sprintf("config err: %v\n%+v", key, err))
	}
	return u
}

// 获取float64类型的配置项,不存在配置项则返回零值和NotFoundErr
func (c *Config) GetFloat64(key string) (float64, error) {
	s, err := c.GetString(key)
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "prop value: %s with key: %s is not float value", s, key)
	}
	return v, nil
}

// 获取float64类型的配置项,不存在配置项则返回第二个参数并且error==nil
func (c *Config) GetFloat64Default(key string, val float64) (float64, error) {
	f, err := c.GetFloat64(key)
	if err != nil && NotFound(err) {
		return val, nil
	}
	return f, err
}

// 获取float64类型的配置项，与GetFloat64一样，除了遇到错误会panic
func (c *Config) MustGetFloat64(key string) float64 {
	f, err := c.GetFloat64(key)
	if err != nil {
		panic(fmt.Sprintf("config err: %v\n%+v", key, err))
	}
	return f
}

// 获取[]string类型的配置项，不存在配置项则返回零值和NotFoundErr
func (c *Config) GetSliceString(key string) ([]string, error) {
	var v []string
	err := c.Get(key, &v)
	return v, err
}

// 获取[]string类型的配置项，与GetSliceString一样，除了遇到错误会panic
func (c *Config) MustGetSliceString(key string) []string {
	s, err := c.GetSliceString(key)
	if err != nil {
		panic(fmt.Sprintf("config err: %v\n%+v", key, err))
	}
	return s
}

// 获取[]int类型的配置项，不存在配置项则返回零值和NotFoundErr
func (c *Config) GetSliceInt(key string) ([]int, error) {
	var v []int
	err := c.Get(key, &v)
	return v, err
}

// 回去[]int类型的配置项，与GetSliceInt一样，除了遇到错误会panic
func (c *Config) MustGetSliceInt(key string) []int {
	i, err := c.GetSliceInt(key)
	if err != nil {
		panic(fmt.Sprintf("config err: %v\n%+v", key, err))
	}
	return i
}

func (c *Config) Get(key string, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.Errorf("参数必须是指针: %T", v)
	}
	return c.get(key, rv.Elem())
}

func (c *Config) get(key string, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Bool:
		return c.getBool(key, v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return c.getInt(key, v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return c.getUint(key, v)
	case reflect.Float32, reflect.Float64:
		return c.getFloat(key, v)
	case reflect.String:
		return c.getString(key, v)
	case reflect.Ptr:
		return c.getPtr(key, v)
	case reflect.Interface:
		return c.getInterface(key, v)
	case reflect.Array:
		return c.getArray(key, v)
	case reflect.Slice:
		return c.getSlice(key, v)
	case reflect.Map:
		return c.getMap(key, v)
	case reflect.Struct:
		return c.getStruct(key, v)
	default:
		return errors.Errorf("不支持的类型: key: %v, value type: %v", key, v)
	}
}

func (c *Config) getBool(key string, v reflect.Value) error {
	b, err := c.GetBool(key)
	if err != nil {
		return err
	}
	v.SetBool(b)
	return nil
}

func (c *Config) getInt(key string, v reflect.Value) error {
	i, err := c.GetInt64(key)
	if err != nil {
		return err
	}
	v.SetInt(i)
	return nil
}

func (c *Config) getUint(key string, v reflect.Value) error {
	u, err := c.GetUint64(key)
	if err != nil {
		return err
	}
	v.SetUint(u)
	return nil
}

func (c *Config) getFloat(key string, v reflect.Value) error {
	f, err := c.GetFloat64(key)
	if err != nil {
		return err
	}
	v.SetFloat(f)
	return nil
}

func (c *Config) getString(key string, v reflect.Value) error {
	s, err := c.GetString(key)
	if err != nil {
		return err
	}
	v.SetString(s)
	return nil
}

func (c *Config) getArray(key string, v reflect.Value) error {
	// 长度为0的数组,不用获取值
	len := v.Len()
	if len == 0 {
		return nil
	}

	var rerr error = NotFoundErr{key: key}
	for i := 0; i < len; i++ {
		err := c.get(fmt.Sprintf("%s[%d]", key, i), v.Index(i))
		if err != nil && !NotFound(err) {
			return err
		}
		if err == nil {
			rerr = nil
		}
	}
	return rerr
}

func (c *Config) getInterface(key string, v reflect.Value) error {
	if v.IsNil() {
		s := ""
		sv := reflect.ValueOf(&s).Elem()
		if sv.Type().AssignableTo(v.Type()) {
			err := c.getString(key, sv)
			if err != nil {
				return err
			}
			v.Set(sv)
		}
		return nil
	} else {
		return c.get(key, v.Elem())
	}
}

func (c *Config) getPtr(key string, v reflect.Value) error {
	if v.IsNil() {
		pv := reflect.New(v.Type().Elem())
		err := c.get(key, pv.Elem())
		if err != nil {
			return err
		}
		v.Set(pv)
		return nil
	} else {
		return c.get(key, v)
	}
}

func (c *Config) getSlice(key string, v reflect.Value) error {
	// 如果slice是nil创建一个长度0的slice替换原来的slice
	// 否则重置slice
	sv := v
	if v.IsNil() {
		sv = reflect.MakeSlice(v.Type(), 0, 3)
	} else {
		// 重置
		sv = v.Slice(0, 0)
	}

	eleTyp := v.Type().Elem()
	var nfe error = NotFoundErr{key: key}
	for i := 0; ; i++ {
		ele := reflect.New(eleTyp).Elem()
		err := c.get(fmt.Sprintf("%s[%d]", key, i), ele)
		if err != nil {
			if NotFound(err) {
				break
			}
			return err
		}
		nfe = nil
		sv = reflect.Append(sv, ele)
	}

	// 设置slice新的值, 如果原来的slice中有值但没有找到配置则使用重置的slice
	if nfe == nil || !v.IsNil() {
		v.Set(sv)
	}

	return nfe
}

func (c *Config) getMap(key string, v reflect.Value) error {
	return errors.New("目前不支持map")
}

func (c *Config) getStruct(key string, v reflect.Value) error {
	var nfe error = NotFoundErr{key: key}

	typ := v.Type()
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		fv := v.Field(i)

		// 小写字段无法设置
		if !fv.CanSet() {
			continue
		}

		k := key
		if !f.Anonymous {
			prop := f.Tag.Get("conf")
			// 跳过此字段
			if prop == "-" {
				continue
			}
			if prop == "" {
				s := f.Name
				prop = strings.ToLower(s[:1]) + s[1:]
			}
			if k != "" {
				k += "."
			}
			k += prop
		}
		err := c.get(k, fv)
		if err != nil && !NotFound(err) {
			return err
		}
		if err == nil {
			nfe = nil
		}
	}

	return nfe
}

// NewConfig 创建配置
func NewConfig() *Config {
	return &Config{}
}
