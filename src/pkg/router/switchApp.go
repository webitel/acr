/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package router

import (
	"github.com/webitel/acr/src/pkg/logger"
	"gopkg.in/mgo.v2/bson"
)

type SwitchApp struct {
	baseApp
	variable string
	cases    map[string]*Node
}

func (c *SwitchApp) Execute(i *Iterator) {
	var ok bool
	var newNode *Node

	if newNode, ok = c.cases[i.Call.ParseString(c.variable)]; ok {
		newNode.setFirst()
		i.SetRoot(newNode)
		logger.Debug("Call %s set switch case: %s", i.Call.GetUuid(), c.variable)
	} else if newNode, ok = c.cases["default"]; ok {
		newNode.setFirst()
		i.SetRoot(newNode)
		logger.Debug("Call %s set switch default case %s", i.Call.GetUuid(), c.variable)
	}
}

func (c *SwitchApp) initConfig(params bson.M) (err error) {
	return nil
}

func NewSwitchApplication(conf AppConfig, parent *Node) *SwitchApp {
	c := &SwitchApp{}
	c.name = "switch"
	c.cases = make(map[string]*Node)
	c.setAppConfig(conf)
	c.setParentNode(parent)
	return c
}
