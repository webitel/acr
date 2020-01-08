/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/router"
)

func ExecuteFunction(scope Scope, c *Call, args interface{}) error {
	var name string
	var ok bool
	var iterator *router.Iterator

	if name, ok = args.(string); ok && name != "" {
		//FIXME add scope
		if iterator, ok = c.Iterator().Functions[name]; ok {
			c.LogDebug("execute", name, "start")
			oldIter := c.Iterator()
			c.SetIterator(iterator)
			c.iterateCallApplication(iterator)
			c.SetIterator(oldIter)

		} else {
			c.LogError("execute", name, "not found")
		}
	} else {
		c.LogError("execute", args, "bad request")
	}

	return nil
}
