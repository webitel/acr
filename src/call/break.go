/**
 * Created by I. Navrotskyj on 21.08.17.
 */

package call

func Break(scope Scope, c *Call, args interface{}) error {
	c.SetBreak()
	c.LogDebug("break", "", "successful")
	return nil
}
