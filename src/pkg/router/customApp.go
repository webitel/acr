/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package router

import (
	"fmt"
)

type CustomApp struct {
	baseApp
	args interface{}
}

func (a *CustomApp) GetArgs() interface{} {
	return a.args
}

func (a *CustomApp) GetType() {

}

func (a *CustomApp) Execute(i *Iterator) {
	fmt.Println("Param: ", a.name, a._depth, a.idx)
	//if execApp, ok := applications.MapApp[a.name]; ok {
	//	execApp()
	//} else {
	//	fmt.Println("Not found app: ", a.name)
	//}
}

func NewCustomApplication(name, id string, conf AppConfig, parent *Node, args interface{}) *CustomApp {
	c := &CustomApp{}
	c.name = name
	c.args = args
	c._id = id
	c.setParentNode(parent)
	c.setAppConfig(conf)
	return c
}
