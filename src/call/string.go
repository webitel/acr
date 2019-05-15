/**
 * Created by I. Navrotskyj on 28.08.17.
 */

package call

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"github.com/robertkrimen/otto"
)

func String(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var vm *otto.Otto
	var varName, fnName, data, value string
	var argsElem []interface{}
	var _args interface{}

	if props, ok = args.(map[string]interface{}); ok {
		if varName = getStringValueFromMap("setVar", props, ""); varName == "" {
			c.LogError("string", props, "setVar is require")
			return nil
		}

		if fnName = getStringValueFromMap("fn", props, ""); fnName == "" {
			c.LogError("string", props, "fn is require")
			return nil
		}

		data = c.ParseString(getStringValueFromMap("data", props, ""))

		switch fnName {
		case "reverse":
			value = reverse(data)
			break
		case "charAt":
			if pos := getIntValueFromMap("args", props, -1); pos > -1 {
				value = charAt(data, pos)
			}
			break
		case "base64":
			mode := ""
			if _args, ok = props["args"]; ok {
				mode = parseInterfaceToString(_args)
			}
			value = base64Fn(mode, data)
			break
		case "MD5":
			value = md5Fn(data)
			break
		case "SHA-256":
			value = sha256Fn(data)
			break
		case "SHA-512":
			value = sha512Fn(data)
			break
		default:
			if _args, ok = props["args"]; ok {
				argsElem = parseArgsToArrayInterface(c, _args)
			} else {
				argsElem = []interface{}{}
			}

			vm = otto.New()
			vm.Set("fnName", fnName)
			vm.Set("args", argsElem)
			vm.Set("data", data)
			v, err := vm.Run(`
				var value, match;

				if (args instanceof Array) {
					args = args.map(function(v) {
						if (typeof v === "string") {
							match = v.match(new RegExp('^/(.*?)/([gimy]*)$'));
							if (match) {
								return new RegExp(match[1], match[2])
							}
						}
						return v;
					})
				} else {
					args = [args]
				}

				if (typeof data[fnName] === "function") {
					value = data[fnName].apply(data, args)
				} else {
					throw "Bad string function " + fnName
				}
			`)

			if err != nil {
				c.LogError("string", props, err.Error())
				return nil
			}

			value = v.String()
		}
		c.LogDebug("string", value, "success")
		return SetVar(c, varName+"="+value)

	} else {
		c.LogError("string", props, "bad request")
	}
	return nil
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func charAt(s string, pos int) string {
	if len(s) > pos {
		return string(s[pos])
	}
	return ""
}

func base64Fn(mode, data string) string {
	if mode == "encoder" {
		return base64.StdEncoding.EncodeToString([]byte(data))
	} else if mode == "decoder" {
		body, _ := base64.StdEncoding.DecodeString(data)
		return string(body)
	}
	return ""
}

func md5Fn(data string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(data)))
}

func sha256Fn(data string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(data)))
}

func sha512Fn(data string) string {
	return fmt.Sprintf("%x", sha512.Sum512([]byte(data)))
}
