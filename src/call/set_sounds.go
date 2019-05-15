/**
 * Created by I. Navrotskyj on 29.08.17.
 */

package call

import (
	"strings"
)

func SetSounds(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var lang, voice string
	var s []string
	var err error

	if props, ok = args.(map[string]interface{}); ok {
		if lang = getStringValueFromMap("lang", props, ""); lang == "" {
			c.LogError("setSounds", props, "lang is require")
			return nil
		}

		if voice = getStringValueFromMap("voice", props, ""); voice == "" {
			c.LogError("setSounds", props, "voice is require")
			return nil
		}

		lang = strings.ToLower(lang)
		s = strings.Split(lang, "_")

		if len(s) < 1 {
			c.LogError("setSounds", props, "bad parse lang")
			return nil
		}

		err = SetVar(c, []string{
			`sound_prefix=/$${sounds_dir}/` + strings.Join(s, `/`) + `/` + voice,
			"default_language=" + s[0],
		})

		if err != nil {
			c.LogError("setSounds", props, err.Error())
			return err
		}
		c.LogDebug("setSounds", props, "success")
	} else {
		c.LogError("setSounds", args, "bad request")
		return nil
	}
	return nil
}
