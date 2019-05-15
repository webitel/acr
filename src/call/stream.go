/**
 * Created by I. Navrotskyj on 24.10.17.
 */

package call

func Stream(c *Call, args interface{}) error {
	var uri = ""
	switch args.(type) {
	case string:
		uri = args.(string)
	default:
		c.LogError("stream", args, "bad request")
		return nil
	}

	err := c.Execute("playback", uri)
	if err != nil {
		c.LogError("stream", uri, err.Error())
		return err
	}
	c.LogError("stream", uri, "success")
	return nil
}
