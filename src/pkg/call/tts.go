/**
 * Created by I. Navrotskyj on 30.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"gopkg.in/mgo.v2/bson"
	"net/url"
	"strconv"
)

func TTS(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var text string

	if props, ok = args.(map[string]interface{}); !ok {
		logger.Error("Call %s tts bad arguments %s", c.Uuid, args)
		return nil
	}

	if text = getStringValueFromMap("text", props, ""); text == "" {
		logger.Error("Call %s tts text is required", c.Uuid)
		return nil
	}

	text = url.QueryEscape(c.ParseString(text))

	switch getStringValueFromMap("provider", props, "") {
	case "microsoft":
		return ttsMicrosoft(c, props, text)
	case "polly":
		return ttsPolly(c, props, text)
	default:
		return ttsDefault(c, props, text)
	}
}

func ttsDefault(c *Call, props map[string]interface{}, text string) error {
	query := "text=" + text
	var key, keyObj string
	var val, valObj interface{}
	var ok bool

	for key, val = range props {
		if key == "text" || key == "getDigits" || key == "broadcast" || key == "terminator" ||
			key == "accessKey" || key == "accessToken" {
			continue
		}

		switch val.(type) {
		case string:
			query += "&" + key + "=" + val.(string)
		case bson.M:
			for keyObj, valObj = range val.(bson.M) {
				if _, ok = valObj.(string); ok {
					query += "&" + keyObj + "=" + valObj.(string)
				}
			}
		}
	}

	return ttsToPlayback(c, props, query, "default")
}

func ttsMicrosoft(c *Call, props map[string]interface{}, text string) error {
	query := "text=" + text
	var ok bool
	var tmp string

	if _, ok = props["voice"]; ok {
		if _, ok = props["voice"].(bson.M); ok {
			voice := props["voice"].(bson.M)

			if tmp = getStringValueFromMap("gender", voice, ""); tmp != "" {
				query += "&gender=" + tmp
			}
			if tmp = getStringValueFromMap("language", voice, ""); tmp != "" {
				query += "&language=" + tmp
			}
			if tmp = getStringValueFromMap("name", voice, ""); tmp != "" {
				query += "&name=" + tmp
			}
		}
	}

	return ttsToPlayback(c, props, query, "microsoft")
}

func ttsPolly(c *Call, props map[string]interface{}, text string) error {
	query := "text=" + text
	var tmp string

	if tmp = getStringValueFromMap("voice", props, ""); tmp != "" {
		query += "&voice=" + tmp
	}

	return ttsToPlayback(c, props, query, "polly")
}

func ttsGetCodecSettings(writeRateVar string) (rate string, format string) {
	rate = "8000"
	format = "mp3"

	if writeRateVar != "" {
		if i, err := strconv.Atoi(writeRateVar); err == nil {
			if i == 8000 || i == 16000 {
				format = ".wav"
				return
			} else if i >= 22050 {
				rate = "22050"
			}
		}
	}
	return
}

func ttsAddCredential(key, token string) string {
	if key != "" && token != "" {
		return "&accessKey=" + url.QueryEscape(key) + "&accessToken=" + url.QueryEscape(token)
	}
	return ""
}

func ttsToPlayback(c *Call, props map[string]interface{}, query, provider string) error {
	var tmp string
	var ok bool
	rate, format := ttsGetCodecSettings(c.GetChannelVar("write_rate"))

	if format == "mp3" {
		tmp = "shout"
	} else {
		tmp = "http_cache"
	}

	playback := map[string]interface{}{
		"name": httpToShot.ReplaceAllString(c.GetGlobalVar("cdr_url"), tmp) + "/sys/tts/" + provider + "?" +
			query +
			ttsAddCredential(getStringValueFromMap("accessKey", props, ""), getStringValueFromMap("accessToken", props, "")) +
			"&rate=" + rate + "&format=" + format,
		"type": "local",
	}

	if _, ok = props["getDigits"]; ok {
		playback["getDigits"] = props["getDigits"]
	}

	if _, ok = props["broadcast"]; ok {
		playback["broadcast"] = props["broadcast"]
	}

	if _, ok = props["terminator"]; ok {
		playback["terminator"] = props["terminator"]
	}

	return Playback(c, playback)
}
