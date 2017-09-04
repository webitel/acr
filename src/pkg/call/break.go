/**
 * Created by I. Navrotskyj on 21.08.17.
 */

package call

func Break(c *Call, args interface{}) error {
	c.SetBreak()
	return nil
}
