package call

import (
	"encoding/json"
	"github.com/webitel/acr/src/model"
	"github.com/webitel/acr/src/router"
	"io/ioutil"
	"net/http"
	"time"
)

func HttpApi(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var res *http.Response
	var req *http.Request
	var err error

	if props, ok = args.(map[string]interface{}); !ok {
		c.LogError("httpRequest", args, "bad request")

		return nil
	}

	req, err = buildRequest(c, props)
	if err != nil {
		c.LogError("httpRequest", args, err.Error())

		return nil
	}

	client := &http.Client{
		Timeout: time.Duration(getIntValueFromMap("timeout", props, 1000)) * time.Millisecond,
	}
	res, err = client.Do(req)
	if err != nil {
		c.LogError("httpRequest", props, err.Error())

		return nil
	}
	defer res.Body.Close()

	var schema model.ArrayApplications
	var data []byte

	if data, err = ioutil.ReadAll(res.Body); err != nil {
		c.LogError("httpRequest", props, err.Error())

		return nil
	}

	if err = json.Unmarshal(data, &schema); err != nil {
		c.LogError("http-api", args, err.Error())

		return nil
	}

	iter := router.NewIterator("http-api", schema, c)
	oldIter := c.Iterator()
	c.LogDebug("httpApi", schema, "switch to new iterator")
	c.SetIterator(iter)
	c.iterateCallApplication(iter)
	c.LogDebug("httpApi", schema, "switch to old iterator")
	c.SetIterator(oldIter)

	return nil
}
