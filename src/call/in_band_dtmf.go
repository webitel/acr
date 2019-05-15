/**
 * Created by I. Navrotskyj on 28.08.17.
 */

package call

func InBandDTMF(c *Call, args interface{}) error {
	var stopDTMF bool
	var app string
	if _, stopDTMF = args.(string); stopDTMF {
		if args.(string) == "stop" {
			stopDTMF = true
		} else {
			stopDTMF = false
		}
	}

	if stopDTMF {
		app = "stop_dtmf"
	} else {
		app = "start_dtmf"
	}

	err := c.Execute(app, "")
	if err != nil {
		c.LogError("inBandDTMF", app, err.Error())
		return err
	}
	c.LogDebug("inBandDTMF", app, "success")
	return nil
}
