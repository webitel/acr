/**
 * Created by I. Navrotskyj on 29.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
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
			logger.Error("Call %s setSounds lang is require", c.Uuid)
			return nil
		}

		if voice = getStringValueFromMap("voice", props, ""); voice == "" {
			logger.Error("Call %s setSounds voice is require", c.Uuid)
			return nil
		}

		lang = strings.ToLower(lang)
		s = strings.Split(lang, "_")

		if len(s) < 1 {
			logger.Error("Call %s setSounds bad parse lang: %s", c.Uuid, lang)
			return nil
		}

		err = SetVar(c, []string{
			`sound_prefix=/$${sounds_dir}/` + strings.Join(s, `/`) + `/` + voice,
			"default_language=" + s[0],
		})

		if err != nil {
			logger.Error("Call %s setSounds error: %s", c.Uuid, err.Error())
			return err
		}

		logger.Debug("Call %s setSounds %s %s successful", c.Uuid, lang, voice)
	} else {
		logger.Error("Call %s setSounds bad arguments: %v", c.Uuid, args)
		return nil
	}
	return nil
}
