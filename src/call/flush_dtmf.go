/**
 * Created by I. Navrotskyj on 28.08.17.
 */

package call

func FlushDTMF(c *Call, args interface{}) error {
	err := c.Execute("flush_dtmf", "")
	if err != nil {
		c.LogError("flushDTMF", nil, err.Error())
		return err
	}
	c.LogDebug("flushDTMF", nil, "success")
	return nil
}
