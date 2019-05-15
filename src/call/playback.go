/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package call

import (
	"fmt"
	"github.com/webitel/acr/src/model"
	"regexp"
	"strings"
)

var httpToShot = regexp.MustCompile(`https?`)

func Playback(c *Call, args interface{}) error {
	var filePath, terminator string
	var ok bool
	var props, getDigits map[string]interface{}

	if props, ok = args.(map[string]interface{}); ok {
		terminator = getStringValueFromMap("terminator", props, "#")
		filePath = playbackGetFileString(c, props)

		if filePath != "" {

			if _, ok = props["getDigits"]; ok {
				if getDigits, ok = applicationToMapInterface(props["getDigits"]); ok {
					return playbackGetDigits(c, getDigits, filePath, terminator)
				}
			} else {
				return playbackSinge(c, filePath, getStringValueFromMap("broadcast", props, ""), terminator)
			}
		}
	} else {
		c.LogError("playback", args, "bad request")
	}
	return nil
}

func playbackGetFileString(c *Call, props map[string]interface{}) (filePath string) {
	var name, typeFile, lang, method string
	var ok, refresh bool
	var files model.ArrayApplications

	name = getStringValueFromMap("name", props, "")
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

	return
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

	err := call.Execute("play_and_get_digits", data)
	if err != nil {
		call.LogError("playback", data, err.Error())
		return err
	}
	call.LogDebug("playback", data, "success")
	return nil
}

func playbackSinge(call *Call, filePath, broadcast, terminator string) (err error) {

	if broadcast != "" {
		if !(broadcast == "aleg" || broadcast == "bleg" || broadcast == "both") {
			broadcast = "both"
		}

		_, err = call.Api(fmt.Sprintf("uuid_broadcast %s %s %s", call.Id(), filePath, broadcast))
		return err
	} else {
		if terminator != "" {
			err = SetVar(call, "playback_terminators="+terminator)
			if err != nil {
				return err
			}
		}

		err := call.Execute("playback", filePath)
		if err != nil {
			call.LogError("playback", filePath, err.Error())
			return err
		}
		call.LogDebug("playback", filePath, "success")
		return nil
	}

}

func getPlaybackFileString(call *Call, typeFile, fileName string, refresh, noPref bool, lang, method string) string {
	var filePath string

	fileName = call.ParseString(fileName)

	switch typeFile {
	case "wav":
		cdrUri := call.GetGlobalVariable("cdr_url")
		if cdrUri != "" {
			if refresh {
				filePath = "{refresh=true}"
			}
			filePath += "http_cache://" + cdrUri + "/sys/media/wav/" + UrlEncoded(fileName) + "?stream=false&domain=" + UrlEncoded(call.Domain()) + "&.wav"
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
		cdrUri := call.GetGlobalVariable("cdr_url")
		if cdrUri != "" {
			filePath = httpToShot.ReplaceAllLiteralString(cdrUri, "shout") + "/sys/media/mp3/" +
				UrlEncoded(fileName) + "?domain=" + UrlEncoded(call.Domain())
		}
	}

	return filePath
}
