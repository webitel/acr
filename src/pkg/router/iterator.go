/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package router

import (
	"github.com/webitel/acr/src/pkg/logger"
	"gopkg.in/mgo.v2/bson"
	"strconv"
)

const MAX_GOTO = 100 //32767

type Tag struct {
	parent *Node
	idx    int
}

type CallFlow struct {
	Id           bson.ObjectId `json:"id" bson:"_id"`
	Name         string        `json:"name" bson:"name"`
	Number       string        `json:"destination_number" bson:"destination_number"`
	Timezone     string        `json:"fs_timezone" bson:"fs_timezone"`
	Domain       string        `json:"domain" bson:"domain"`
	Callflow     []interface{} `json:"callflow" bson:"callflow"`
	OnDisconnect []interface{} `json:"onDisconnect" bson:"onDisconnect"`
	Version      int           `json:"version" bson:"version"`
}

type Iterator struct {
	Call        Call
	Tags        map[string]*Tag
	Functions   map[string]*Iterator
	currentNode *Node
	gotoCounter int16
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
		logger.Warning("Call %s max goto count!", i.Call.GetUuid())
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

func (i *Iterator) parseCallFlowArray(root *Node, cf []interface{}) {
	var ok bool
	var appName, tag string
	var configFlags AppConfig
	var args interface{}

	var tmpMap map[string]interface{}
	var tmp, v interface{}

	for _, v = range cf {

		if _, ok = v.(bson.M); !ok {
			continue
		}

		appName, args, configFlags, tag = parseApp(v.(bson.M), i.Call)
		switch appName {
		case "if":

			if tmpMap, ok = args.(map[string]interface{}); ok {
				condApp := NewConditionApplication(configFlags, root)
				if tmp, ok = tmpMap["then"]; ok {
					if _, ok = tmp.([]interface{}); ok {
						i.parseCallFlowArray(condApp._then, tmp.([]interface{}))
					}
				}

				if tmp, ok = tmpMap["else"]; ok {
					if _, ok = tmp.([]interface{}); ok {
						i.parseCallFlowArray(condApp._else, tmp.([]interface{}))
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
								i.Functions[tmpMap["name"].(string)] = NewIterator(tmpMap["actions"].([]interface{}), i.Call)
								continue
							}
						}
					}
				}
			}
		case "switch":
			if tmpMap, ok = args.(map[string]interface{}); ok {
				switchApp := NewSwitchApplication(configFlags, root)
				i.trySetTag(tag, switchApp, root, switchApp.idx)
				root.Add(switchApp)

				if _, ok = tmpMap["variable"]; ok {
					if _, ok = tmpMap["variable"].(string); ok {
						switchApp.variable = tmpMap["variable"].(string)
					}
				}

				if _, ok = tmpMap["case"]; ok {
					if _, ok = tmpMap["case"].(bson.M); ok {
						for caseName, caseValue := range tmpMap["case"].(bson.M) {
							if _, ok = caseValue.([]interface{}); ok {
								switchApp.cases[caseName] = NewNode(root)
								i.parseCallFlowArray(switchApp.cases[caseName], caseValue.([]interface{}))
							}
						}
					}
				}
			}

		default:
			if appName == "" && configFlags&flagBreakEnabled == flagBreakEnabled {
				appName = "break"
			}
			customApp := NewCustomApplication(appName, configFlags, root, args)
			i.trySetTag(tag, customApp, root, customApp.idx)
			root.Add(customApp)

		}

	}
}

func NewIterator(c []interface{}, call Call) *Iterator {
	i := &Iterator{}
	i.Call = call
	i.currentNode = NewNode(nil)
	i.Functions = make(map[string]*Iterator)
	i.Tags = make(map[string]*Tag)
	i.parseCallFlowArray(i.currentNode, c)
	return i
}

func parseApp(m bson.M, c Call) (appName string, args interface{}, appConf AppConfig, tag string) {

	for fieldName, fieldValue := range m {
		switch fieldName {
		case "break":
			if v, ok := fieldValue.(bool); ok && v {
				appConf |= flagBreakEnabled
			}
		case "async":
			if v, ok := fieldValue.(bool); ok && v {
				appConf |= flagAsyncEnabled
			}
		case "dump":
			if v, ok := fieldValue.(bool); ok && v {
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
			if !(c.ValidateApp(fieldName) || fieldName == "if" || fieldName == "switch" || fieldName == "function") {
				continue
			}

			if appName == "" {
				appName = fieldName

				if m, ok := fieldValue.(bson.M); ok {
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

func init2() {
	//region json
	//const jsonStream = `
	//	[
	//		{
	//			"break": true,
	//			"ddd":  1
	//		}
	//	]
	//`

	//endregion

	//session, err := mgo.Dial("10.10.10.200:27017")
	//if err != nil {
	//	panic(err)
	//}
	//defer session.Close()
	//c := session.DB("webitel").C("default")
	//
	//result := &CallFlow{}
	//
	//err = c.Find(bson.M{"name": "go"}).One(&result)
	//if err != nil {
	//	panic(err)
	//}
	//
	//iter := NewIterator(result, nil)
	//
	//for {
	//	v := iter.NextApp()
	//	if v == nil {
	//		break
	//	}
	//	fmt.Println(v)
	//	v.Execute(iter)
	//}
	//fmt.Println(iter.Tags)

}
