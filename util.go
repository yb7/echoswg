package echoswg

import (
  "bytes"
  "io/ioutil"
  "net/http"
  "strings"
)

type PathNames []string
// PathNames func
func ParsePathNames(path string) PathNames {
	var pnames PathNames = []string{} // Param names
	for i, l := 0, len(path); i < l; i++ {
		if path[i] == ':' {
			j := i + 1

			for ; i < l && path[i] != '/'; i++ {
			}

			pnames = append(pnames, path[j:i])
			path = path[:j] + path[i:]
			i, l = j, len(path)
		} else if path[i] == '*' {
			pnames = append(pnames, "_*")
		}
	}
	return pnames
}

func (pnames PathNames) contains(key string) bool {
	for _, pname := range pnames {
		if pname == key {
			return true
		}
	}
	return false
}

func containsIgnoreCase(s []string, e string) bool {
	for _, a := range s {
		if strings.ToUpper(a) == strings.ToUpper(e) {
			return true
		}
	}
	return false
}



func lowCamelStr(str string) string {
	for _, word := range []string{"ID", "URL", "URI"} {
		if word == str {
			return strings.ToLower(str)
		}
	}
	return strings.ToLower(string(str[0])) + string(str[1:])
}
/**
 * 对外使用，不要删除
 */
func CopyRequestBody(req *http.Request) ([]byte, error) {
  buf, err := ioutil.ReadAll(req.Body)
  if err != nil {
    return nil, err
  }
  req.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
  return buf, nil
}

func mergeMap(inputs ...map[string]any) map[string]any {
  m := make(map[string]any)
  for _, input := range inputs {
    for k, v := range input {
      m[k] = v
    }
  }
  return m
}
