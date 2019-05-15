/**
 * Created by I. Navrotskyj on 21.08.17.
 */

package call

func UnSet(c *Call, args interface{}) error {

	if data, ok := args.(string); ok {
		err := c.Execute("unset", data)
		if err != nil {
			c.LogError("unSet", data, err.Error())
			return err
		}
		c.LogDebug("unSet", data, "success")
	} else {
		c.LogError("unSet", args, "bad request")
	}
	return nil
}
