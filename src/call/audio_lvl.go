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

	data := fmt.Sprintf("%s %s",
		getStringValueFromMap("direction", props, AUDIO_LEVEL_DIRECTION_READ),
		getStringValueFromMap("level", props, "-1"),
	)

	err := c.Execute("set_audio_level", data)
	if err != nil {
		c.LogError("SetAudioLevel", data, err.Error())
		return err
	}
	c.LogDebug("SetAudioLevel", data, "success")
	return nil
}
