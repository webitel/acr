/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

func Sleep(scope Scope, c *Call, args interface{}) error {

	c.LogDebug("sleep", args, "start")
	err := c.Execute("sleep", parseInterfaceToString(args))
	if err != nil {
		c.LogError("sleep", args, err.Error())
		return err
	}
	c.LogDebug("sleep", args, "successful")
	return nil
}
