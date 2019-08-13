package call

import "fmt"

const (
	AUDIO_LEVEL_DIRECTION_READ  = "read"
	AUDIO_LEVEL_DIRECTION_WRITE = "write"
)

func SetAudioLevel(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool

	if props, ok = args.(map[string]interface{}); !ok {
		c.LogError("SetAudioLevel", args, "bad request")
		return nil
	}

	data := fmt.Sprintf("uuid_audio %s start %s %s %s",
		c.Id(),
		getStringValueFromMap("direction", props, AUDIO_LEVEL_DIRECTION_READ),
		getStringValueFromMap("action", props, "level"),
		getStringValueFromMap("level", props, "3"),
	)

	_, err := c.Api(data)
	if err != nil {
		c.LogError("SetAudioLevel", data, err.Error())
		return err
	}
	c.LogDebug("SetAudioLevel", data, "success")
	return nil
}
