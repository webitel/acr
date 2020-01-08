/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package router

import (
	"fmt"
	"github.com/webitel/acr/src/model"
	"github.com/webitel/wlog"
	"strconv"
	"sync"
)

const MAX_GOTO = 100 //32767

type Tag struct {
	parent *Node
	idx    int
}

type Iterator struct {
	name        string
	Call        Call
	Tags        map[string]*Tag
	Functions   map[string]*Iterator
	triggers    map[string]*Iterator
	currentNode *Node
	gotoCounter int16
	cancel      bool
	sync.RWMutex
}

func (i *Iterator) Name() string {
	return i.name
}

func (i *Iterator) SetRoot(root *Node) {
	i.currentNode = root
}

func (i *Iterator) NextApp() App {
	var app App
	app = i.currentNode.Next()
	if app == nil {
		if newNode := i.GetParentNode(); newNode == nil {
			return nil
		} else {
			return i.NextApp()
		}
	} else {
		return app
	}
}

func (i *Iterator) GetParentNode() *Node {
	parent := i.currentNode.GetParent()
	i.currentNode.setFirst()
	if parent == nil {
		return nil
	}
	i.currentNode = parent
	return i.currentNode
}

func (i *Iterator) trySetTag(tag string, a App, parent *Node, idx int) {
	if tag != "" {
		i.Tags[tag] = &Tag{
			parent: parent,
			idx:    idx,
		}
	}
}

func (i *Iterator) Goto(tag string) bool {
	if i.gotoCounter > MAX_GOTO {
		wlog.Warn(fmt.Sprintf("call %s max goto count!", i.Call.Id()))
		return false
	}

	if gotoApp, ok := i.Tags[tag]; ok {
		i.currentNode.setFirst()
		i.SetRoot(gotoApp.parent)
		i.currentNode.position = gotoApp.idx
		if i.currentNode.parent != nil {
			i.currentNode.parent.position = i.currentNode.idx + 1
		}
		i.gotoCounter++
		return true
	}
	return false
}

func (i *Iterator) SetCancel() {
	i.Lock()
	defer i.Unlock()
	i.cancel = true
}

func (i *Iterator) IsCancel() bool {
	i.RLock()
	defer i.RUnlock()
	return i.cancel
}

func (i *Iterator) parseCallFlowArray(root *Node, cf model.ArrayApplications) {
	var ok bool
	var appName, tag, id string
	var configFlags AppConfig
	var args interface{}

	var tmpMap map[string]interface{}
	var tmp, v interface{}

	for _, v = range cf {

		if _, ok = v.(model.Application); !ok {
			continue
		}

		appName, args, configFlags, tag, id = parseApp(v.(model.Application), i.Call)
		switch appName {
		case "if":

			if tmpMap, ok = args.(map[string]interface{}); ok {
				condApp := NewConditionApplication(id, configFlags, root)
				if tmp, ok = tmpMap["then"]; ok {
					if _, ok = tmp.([]interface{}); ok {
						i.parseCallFlowArray(condApp._then, ArrInterfaceToArrayApplication(tmp.([]interface{})))
					}
				}

				if tmp, ok = tmpMap["else"]; ok {
					if _, ok = tmp.([]interface{}); ok {
						i.parseCallFlowArray(condApp._else, ArrInterfaceToArrayApplication(tmp.([]interface{})))
					}
				}

				if tmp, ok = tmpMap["sysExpression"]; ok {
					if _, ok = tmp.(string); ok {
						condApp.expression = tmp.(string)
					}
				}

				i.trySetTag(tag, condApp, root, condApp.idx)
				root.Add(condApp)
			}

		case "function":
			if tmpMap, ok = args.(map[string]interface{}); ok {
				if _, ok = tmpMap["name"]; ok {
					if _, ok = tmpMap["name"].(string); ok {
						if _, ok = tmpMap["actions"]; ok {
							if _, ok = tmpMap["actions"].([]interface{}); ok {
								i.Functions[tmpMap["name"].(string)] = NewIterator("function", ArrInterfaceToArrayApplication(tmpMap["actions"].([]interface{})), i.Call)
								continue
							}
						}
					}
				}
			}
		case "trigger":
			if tmpMap, ok = args.(map[string]interface{}); ok {
				for k, v := range tmpMap {
					if _, ok = v.([]interface{}); ok {
						i.addTrigger(k, v.([]interface{}))
					}
				}
			}
		case "switch":
			if tmpMap, ok = args.(map[string]interface{}); ok {
				switchApp := NewSwitchApplication(id, configFlags, root)
				i.trySetTag(tag, switchApp, root, switchApp.idx)
				root.Add(switchApp)

				if _, ok = tmpMap["variable"]; ok {
					if _, ok = tmpMap["variable"].(string); ok {
						switchApp.variable = tmpMap["variable"].(string)
					}
				}

				if _, ok = tmpMap["case"]; ok {
					if _, ok = tmpMap["case"].(map[string]interface{}); ok {
						for caseName, caseValue := range tmpMap["case"].(map[string]interface{}) {
							if _, ok = caseValue.([]interface{}); ok {
								switchApp.cases[caseName] = NewNode(root)
								i.parseCallFlowArray(switchApp.cases[caseName], ArrInterfaceToArrayApplication(caseValue.([]interface{})))
							}
						}
					}
				}
			}

		default:
			if appName == "" && configFlags&flagBreakEnabled == flagBreakEnabled {
				appName = "break"
			}
			customApp := NewCustomApplication(appName, id, configFlags, root, args)
			i.trySetTag(tag, customApp, root, customApp.idx)
			root.Add(customApp)

		}

	}
}

func NewIterator(name string, c model.ArrayApplications, call Call) *Iterator {
	i := &Iterator{}
	i.name = name
	i.Call = call
	i.currentNode = NewNode(nil)
	i.Functions = make(map[string]*Iterator)
	i.triggers = make(map[string]*Iterator)
	i.Tags = make(map[string]*Tag)
	i.parseCallFlowArray(i.currentNode, c)
	return i
}

func parseApp(m model.Application, c Call) (appName string, args interface{}, appConf AppConfig, tag, _id string) {
	var ok, v bool

	for fieldName, fieldValue := range m {
		switch fieldName {
		case "_id":
			if _, ok = fieldValue.(string); ok {
				_id = fieldValue.(string)
			}
		case "break":
			if v, ok = fieldValue.(bool); ok && v {
				appConf |= flagBreakEnabled
			}
		case "async":
			if v, ok = fieldValue.(bool); ok && v {
				appConf |= flagAsyncEnabled
			}
		case "dump":
			if v, ok = fieldValue.(bool); ok && v {
				appConf |= flagDumpEnabled
			}
		case "tag":
			switch fieldValue.(type) {
			case string:
				tag = fieldValue.(string)
			case int:
				tag = strconv.Itoa(fieldValue.(int))
			}
		default:
			if !(c.ValidateApp(fieldName) || fieldName == "if" || fieldName == "switch" || fieldName == "function" || fieldName == "trigger") {
				continue
			}

			if appName == "" {
				appName = fieldName

				if m, ok := fieldValue.(model.Application); ok {
					tmp := make(map[string]interface{})
					for argK, argV := range m {
						tmp[argK] = argV
					}
					args = tmp
				} else {
					args = fieldValue
				}
			}
		}

	}
	return
}

func ArrInterfaceToArrayApplication(src []interface{}) model.ArrayApplications {
	res := make(model.ArrayApplications, len(src))
	var ok bool
	for k, v := range src {
		if _, ok = v.(map[string]interface{}); ok {
			res[k] = v.(map[string]interface{})
		}
	}
	return res
}
