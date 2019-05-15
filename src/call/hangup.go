/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

func Hangup(c *Call, args interface{}) error {
	switch args.(type) {
	case string:
		return hangupChannel(c, args.(string))
	case map[string]interface{}:
		params := args.(map[string]interface{})
		uuid := c.ParseString(getStringValueFromMap("uuid", params, ""))
		if uuid == "" {
			c.LogError("hangup", args, "require uuid")
			return nil
		}
		return hangupByUuid(
			c,
			uuid,
			c.ParseString(getStringValueFromMap("cause", params, "")),
		)

	default:
		c.LogError("hangup", args, "bad request")
		return nil
	}
	return nil
}

func hangupChannel(c *Call, cause string) error {
	err := c.Hangup(cause)
	if err != nil {
		c.LogError("hangup", cause, err.Error())
		return err
	}
	c.LogDebug("hangup", cause, "success")
	return nil
}

func hangupByUuid(c *Call, uuid, cause string) error {
	_, err := c.Api("uuid_kill " + uuid + " " + cause)
	if err != nil {
		c.LogError("hangup", uuid+" "+cause, err.Error())
	}
	c.LogDebug("hangup", uuid+" "+cause, "successful")
	return nil
}
