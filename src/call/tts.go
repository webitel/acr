/**
 * Created by I. Navrotskyj on 30.08.17.
 */

package call

import (
	"strconv"
)

func TTS(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var text string

	if props, ok = args.(map[string]interface{}); !ok {
		c.LogError("tts", args, "bad request")
		return nil
	}

	if text = getStringValueFromMap("text", props, ""); text == "" {
		c.LogError("tts", args, "text is require")
		return nil
	}

	text = UrlEncoded(c.ParseString(text))

	switch getStringValueFromMap("provider", props, "") {
	case "microsoft":
		return ttsMicrosoft(c, props, text)
	case "polly":
		return ttsPolly(c, props, text)
	case "google":
		return ttsGoogle(c, props, text)
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
		case map[string]interface{}:
			for keyObj, valObj = range val.(map[string]interface{}) {
				if _, ok = valObj.(string); ok {
					query += "&" + keyObj + "=" + valObj.(string)
				}
			}
		}
	}

	return ttsToPlayback(c, props, query, "default", nil)
}

func ttsMicrosoft(c *Call, props map[string]interface{}, text string) error {
	query := "text=" + text
	var ok bool
	var tmp string

	if _, ok = props["voice"]; ok {
		if _, ok = props["voice"].(map[string]interface{}); ok {
			voice := props["voice"].(map[string]interface{})

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

	if tmp = getStringValueFromMap("region", props, ""); tmp != "" {
		query += "&region=" + tmp
	}

	return ttsToPlayback(c, props, query, "microsoft", nil)
}

var mp3Mod = ".mp3"
var wavMod = ".wav"

func ttsGoogle(c *Call, props map[string]interface{}, text string) error {
	query := "text=" + text
	var ok bool
	var tmp string

	var format *string

	if _, ok = props["voice"]; ok {
		if _, ok = props["voice"].(map[string]interface{}); ok {
			voice := props["voice"].(map[string]interface{})

			if tmp = getStringValueFromMap("gender", voice, ""); tmp != "" {
				query += "&gender=" + tmp
			}
			if tmp = getStringValueFromMap("language", voice, ""); tmp != "" {
				query += "&language=" + tmp
			}
			if tmp = getStringValueFromMap("name", voice, ""); tmp != "" {
				query += "&name=" + tmp
			}

			if tmp = getStringValueFromMap("audioEncoding", voice, ""); tmp != "" {
				query += "&audioEncoding=" + tmp

				switch tmp {
				case "OGG_OPUS", "MP3":
					format = &mp3Mod
				default:
					format = &wavMod
				}
			}
			if tmp = getStringValueFromMap("sampleRateHertz", voice, ""); tmp != "" {
				query += "&sampleRateHertz=" + tmp
			}
			if tmp = getStringValueFromMap("speakingRate", voice, ""); tmp != "" {
				query += "&speakingRate=" + tmp
			}
			if tmp = getStringValueFromMap("pitch", voice, ""); tmp != "" {
				query += "&pitch=" + tmp
			}
			if tmp = getStringValueFromMap("volumeGainDb", voice, ""); tmp != "" {
				query += "&volumeGainDb=" + tmp
			}
			if tmp = getStringValueFromMap("effectsProfileId", voice, ""); tmp != "" {
				query += "&effectsProfileId=" + tmp
			}

		}
	}

	if tmp = getStringValueFromMap("textType", props, ""); tmp != "" {
		query += "&textType=" + tmp
	}

	return ttsToPlayback(c, props, query, "google", format)
}

func ttsPolly(c *Call, props map[string]interface{}, text string) error {
	query := "text=" + text
	var tmp string

	if tmp = getStringValueFromMap("voice", props, ""); tmp != "" {
		query += "&voice=" + tmp
	}

	if tmp = getStringValueFromMap("textType", props, ""); tmp != "" {
		query += "&textType=" + tmp
	}

	return ttsToPlayback(c, props, query, "polly", nil)
}

func ttsGetCodecSettings(writeRateVar string, defFormat *string) (rate string, format string) {
	rate = "8000"
	format = "mp3"

	if defFormat != nil {
		format = *defFormat
		return
	}

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
		return "&accessKey=" + UrlEncoded(key) + "&accessToken=" + UrlEncoded(token)
	}
	return ""
}

func ttsToPlayback(c *Call, props map[string]interface{}, query, provider string, defFormat *string) error {
	var tmp string
	var ok bool
	rate, format := ttsGetCodecSettings(c.GetVariable("write_rate"), defFormat)

	if format == "mp3" {
		tmp = "shout"
	} else {
		tmp = "http_cache"
	}

	playback := map[string]interface{}{
		"name": httpToShot.ReplaceAllString(c.GetGlobalVariable("cdr_url"), tmp) + "/sys/tts/" + provider + "?" +
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
