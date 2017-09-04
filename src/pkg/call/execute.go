/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"github.com/webitel/acr/src/pkg/router"
)

func ExecuteFunction(c *Call, args interface{}) error {
	var name string
	var ok bool
	var iterator *router.Iterator

	if name, ok = args.(string); ok && name != "" {
		if iterator, ok = c.Iterator.Functions[name]; ok {
			logger.Debug("Call %s execute function %s", c.Uuid, name)
			oldIter := c.Iterator
			c.Iterator = iterator
			routeIterator(c)
			c.Iterator = oldIter

		} else {
			logger.Error("Call %s execute not found function %s", c.Uuid, name)
		}
	} else {
		logger.Error("Call %s execute function bad arguments %s", c.Uuid, args)
	}

	return nil
}
