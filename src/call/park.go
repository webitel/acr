/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"strings"
)

func Park(c *Call, args interface{}) error {
	var data map[string]interface{}
	var ok bool
	var name, lot, auto string

	if data, ok = args.(map[string]interface{}); ok {
		if _, ok = data["name"]; !ok {
			c.LogError("park", data, "name is required")
			return nil
		}

		if _, ok = data["lot"]; !ok {
			c.LogError("park", data, "lot is required")
			return nil
		}

		if name, ok = data["name"].(string); !ok || name == "" {
			c.LogError("park", data, "name must string")
			return nil
		}

		if lot, ok = data["lot"].(string); !ok || lot == "" {
			c.LogError("park", data, "lot must string")
			return nil
		}

		if _, ok = data["auto"]; ok {
			if auto, ok = data["auto"].(string); ok {
				auto = "auto " + auto + " "
				lot = strings.Replace(lot, "-", " ", -1)
			}
		}

		str := name + "@${domain_name} " + auto + lot
		err := c.Execute("valet_park", str)
		if err != nil {
			c.LogError("park", str, err.Error())
			return err
		}
		c.LogDebug("park", str, "success")
	} else {
		c.LogError("park", args, "bad request")
	}
	return nil
}
