/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package router

import (
	"fmt"
	"github.com/webitel/wlog"
)

type SwitchApp struct {
	baseApp
	variable string
	cases    map[string]*Node
}

func (a *SwitchApp) IteratorApp() bool {
	return true
}

func (c *SwitchApp) Execute(i *Iterator) {
	var ok bool
	var newNode *Node

	if newNode, ok = c.cases[i.Call.ParseString(c.variable)]; ok {
		newNode.setFirst()
		i.SetRoot(newNode)
		wlog.Debug(fmt.Sprintf("call %s set switch case: %s", i.Call.Id(), c.variable))
	} else if newNode, ok = c.cases["default"]; ok {
		newNode.setFirst()
		i.SetRoot(newNode)
		wlog.Debug(fmt.Sprintf("call %s set switch default case %s", i.Call.Id(), c.variable))
	}
}

func (c *SwitchApp) initConfig(params interface{}) (err error) {
	return nil
}

func NewSwitchApplication(id string, conf AppConfig, parent *Node) *SwitchApp {
	c := &SwitchApp{}
	c.name = "switch"
	c._id = id
	c.cases = make(map[string]*Node)
	c.setAppConfig(conf)
	c.setParentNode(parent)
	return c
}
