/**
 * Created by I. Navrotskyj on 21.08.17.
 */

package call

func IVR(c *Call, args interface{}) error {

	if data, ok := args.(string); ok {
		err := c.Execute("ivr", data+"@"+c.Domain())
		if err != nil {
			c.LogError("ivr", data, err.Error())
			return err
		}
		c.LogDebug("ivr", data, "success")
	} else {
		c.LogError("ivr", args, "bad request")
	}
	return nil
}
