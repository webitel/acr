/**
 * Created by I. Navrotskyj on 28.08.17.
 */

package call

//todo need test
func Pickup(scope Scope, c *Call, args interface{}) error {

	if data, ok := args.(string); ok && data != "" {
		err := c.Execute("pickup", data+"@"+c.Domain())
		if err != nil {
			c.LogError("pickup", data, err.Error())
			return err
		}
		c.LogDebug("pickup", data, "success")
	} else {
		c.LogError("pickup", args, "bad request")
	}
	return nil
}
