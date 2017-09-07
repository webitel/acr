/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"gopkg.in/mgo.v2/bson"
	"net/url"
	"regexp"
	"strings"
)

var httpToShot = regexp.MustCompile(`https?`)

func Playback(c *Call, args interface{}) error {
	var filePath, name, typeFile, lang, method, terminator string
	var ok, refresh bool
	var props, getDigits map[string]interface{}
	var files []bson.M

	if props, ok = args.(map[string]interface{}); ok {
		name = getStringValueFromMap("name", props, "")
		terminator = getStringValueFromMap("terminator", props, "#")

		if name != "" {
			typeFile = getStringValueFromMap("type", props, "")
			lang = getStringValueFromMap("lang", props, "")
			method = getStringValueFromMap("method", props, "")

			refresh = false

			if _, ok = props["refresh"]; ok {
				refresh, _ = props["refresh"].(bool)
			}

			filePath = getPlaybackFileString(c, typeFile, name, refresh, false, lang, method)

		} else if _, ok = props["files"]; ok {
			if files, ok = getArrayFromMap(props["files"]); ok {
				for _, file := range files {
					name = getStringValueFromMap("name", file, "")
					typeFile = getStringValueFromMap("type", file, "")
					lang = getStringValueFromMap("lang", file, "")
					method = getStringValueFromMap("method", file, "")
					refresh = false
					if _, ok = props["refresh"]; ok {
						refresh, _ = props["refresh"].(bool)
					}
					filePath += "!" + getPlaybackFileString(c, typeFile, name, refresh, false, lang, method)
				}
				if len(filePath) > 0 {
					filePath = "file_string://" + filePath[1:]
				}
			}
		}

		if filePath != "" {

			if _, ok = props["getDigits"]; ok {
				if getDigits, ok = bsonToMapInterface(props["getDigits"]); ok {
					return playbackGetDigits(c, getDigits, filePath, terminator)
				}
			} else {
				return playbackSinge(c, filePath, getStringValueFromMap("broadcast", props, ""), terminator)
			}
		}

	}

	logger.Error("Call %s playback bad arguments %s", c.Uuid, args)
	return nil
}

func playbackGetDigits(call *Call, getDigitProps map[string]interface{}, filePath, terminator string) error {

	data := strings.Join([]string{
		getStringValueFromMap("min", getDigitProps, "1"),
		getStringValueFromMap("max", getDigitProps, "1"),
		getStringValueFromMap("tries", getDigitProps, "1"),
		getStringValueFromMap("timeout", getDigitProps, "3000"),
		terminator,
		filePath,
		"silence_stream://250",
		getStringValueFromMap("setVar", getDigitProps, "MyVar"),
		getStringValueFromMap("regexp", getDigitProps, ".*"),
	}, " ")

	_, err := call.SndMsg("play_and_get_digits", data, true, true)
	if err != nil {
		logger.Error("Call %s playback getDigits error: %s", call.Uuid, err.Error())
		return err
	}
	logger.Debug("Call %s playback getDigits %s successful", call.Uuid, data)
	return nil
}

func playbackSinge(call *Call, filePath, broadcast, terminator string) (err error) {

	if broadcast != "" {
		if !(broadcast == "aleg" || broadcast == "bleg" || broadcast == "both") {
			broadcast = "both"
		}

		_, err = call.Conn.BgApi("uuid_broadcast", call.Uuid, filePath, broadcast)
		return err
	} else {
		if terminator != "" {
			err = SetVar(call, "playback_terminators="+terminator)
			if err != nil {
				return err
			}
		}

		_, err := call.SndMsg("playback", filePath, true, true)
		if err != nil {
			logger.Error("Call %s playback error: %s", call.Uuid, err.Error())
			return err
		}
		logger.Debug("Call %s playback %s successful", call.Uuid, filePath)
		return nil
	}

}

func getPlaybackFileString(call *Call, typeFile, fileName string, refresh, noPref bool, lang, method string) string {
	var filePath string

	fileName = call.ParseString(fileName)

	switch typeFile {
	case "wav":
		cdrUri := call.GetGlobalVar("cdr_url")
		if cdrUri != "" {
			if refresh {
				filePath = "{refresh=true}"
			}
			filePath += "http_cache://" + cdrUri + "/sys/media/wav/" + url.QueryEscape(fileName) + "?stream=false&domain=" + url.QueryEscape(call.Domain) + "&.wav"
		}

	case "silence":
		if noPref {
			filePath = typeFile
		} else {
			filePath = "silence_stream://" + fileName
		}
	case "local":
		filePath = fileName

	case "shout":
		filePath = httpToShot.ReplaceAllLiteralString(fileName, "shout")

	case "tone":
		if noPref {
			filePath = fileName
		} else {
			filePath = "tone_stream://" + fileName
		}

	case "say":
		filePath = "${say_string " + lang + " " + lang + " " + method + " " + fileName + "}"

	default:
		cdrUri := call.GetGlobalVar("cdr_url")
		if cdrUri != "" {
			filePath = httpToShot.ReplaceAllLiteralString(cdrUri, "shout") + "/sys/media/mp3/" +
				url.QueryEscape(fileName) + "?domain=" + url.QueryEscape(call.Domain)
		}
	}

	return filePath
}
