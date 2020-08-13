package conf

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
)

type Source interface {
	Name() string
	Get(key string) string
}

type MapSource struct {
	name  string
	items map[string]string
}

func (s *MapSource) Name() string {
	return s.name
}

func (s *MapSource) Get(key string) string {
	for k, v := range s.items {
		if k == key {
			return v
		}
	}
	return ""
}

// 创建YAML的配置源
func NewYAMLSource(name string, data []byte) (Source, error) {
	var mapSlice yaml.MapSlice
	err := yaml.Unmarshal(data, &mapSlice)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	source := &MapSource{name: name, items: map[string]string{}}

	addEntry(source.items, "", mapSlice)

	return source, nil
}

func addEntry(entries map[string]string, keyPrefix string, v interface{}) {
	switch v.(type) {
	case yaml.MapSlice:
		if keyPrefix != "" {
			keyPrefix += "."
		}
		for _, m := range v.(yaml.MapSlice) {
			addEntry(entries, fmt.Sprintf("%s%v", keyPrefix, m.Key), m.Value)
		}
	case []interface{}:
		for i, s := range v.([]interface{}) {
			addEntry(entries, fmt.Sprintf("%s[%d]", keyPrefix, i), s)
		}
	default:
		entries[keyPrefix] = fmt.Sprintf("%v", v)
	}
}

// 创建文件配置源
func FileSource(file string) (Source, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Wrap(err, "创建文件配置错误")
	}
	return NewYAMLSource("file:"+file, data)
}

// 创建apollo配置源
func ApolloSource(server, app, env, ns string) (Source, error) {
	config, err := apolloConfig(server, app, env, ns)
	if err != nil {
		return nil, err
	}
	return NewYAMLSource(fmt.Sprintf("apollo-%v-%v-%v", app, env, ns), []byte(config))
}

func apolloConfig(server, app, env, ns string) (string, error) {
	url := fmt.Sprintf(server+"/configfiles/json/%s/%s/%s", app, env, ns)
	resp, err := http.Get(url)
	if err != nil {
		return "", errors.Wrapf(err, "apollo配置获取错误, url: %v", url)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", errors.Errorf("apollo配置获取错误, url: %v, http status code: %v", url, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "apollo配置读取错误")
	}

	var result map[string]interface{}
	if err = json.Unmarshal(body, &result); err != nil {
		return "", errors.Wrap(err, "apollo配置json解析错误")
	}

	if content, ok := result["content"]; ok {
		if s, ok := content.(string); ok {
			return s, nil
		}
		return "", errors.New("apollo配置json解析错误, content属性错误")
	} else {
		if msg, ok := result["message"]; ok {
			return "", errors.Errorf("apollo配置获取错误: %v", msg)
		}
		return "", errors.Errorf("apollo配置获取错误, 内容: %v", result)
	}
}

// 创建命令行配置
func CMDLineSource() Source {
	items := map[string]string{}
	key := ""

	startKey := func(exp string) string {
		// 忽略前面的-
		i := 0
		for i = 0; i < len(exp); i++ {
			if exp[i] != '-' {
				break
			}
		}
		exp = exp[i:]
		for i = 0; i < len(exp); i++ {
			if exp[i] == '=' {
				break
			}
		}
		key = exp[:i]
		return exp[i:]
	}

	startValue := func(exp string) {
		if key != "" {
			items[key] = exp
			key = ""
		}
	}

	startArg := func(exp string) {
		if exp[0] == '-' {
			if key != "" {
				items[key] = "true"
				key = ""
			}
			exp = startKey(exp[1:])
			if len(exp) > 0 {
				startValue(exp[1:])
			}
		} else {
			startValue(exp)
		}
	}

	for _, arg := range os.Args[1:] {
		startArg(arg)
	}

	if key != "" {
		items[key] = "true"
	}

	return &MapSource{name: "cmd", items: items}
}
