/**
 * Created by I. Navrotskyj on 22.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"strings"
)

func AMD(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var tmp string
	data := make([]string, 9)

	if props, ok = args.(map[string]interface{}); ok {

		if tmp = getStringValueFromMap("silenceThreshold", props, ""); tmp != "" {
			data = append(data, "silence_threshold="+tmp)
		}

		if tmp = getStringValueFromMap("maximumWordLength", props, ""); tmp != "" {
			data = append(data, "maximum_word_length="+tmp)
		}

		if tmp = getStringValueFromMap("maximumNumberOfWords", props, ""); tmp != "" {
			data = append(data, "maximum_number_of_words="+tmp)
		}

		if tmp = getStringValueFromMap("betweenWordsSilence", props, ""); tmp != "" {
			data = append(data, "between_words_silence="+tmp)
		}

		if tmp = getStringValueFromMap("minWordLength", props, ""); tmp != "" {
			data = append(data, "min_word_length="+tmp)
		}

		if tmp = getStringValueFromMap("totalAnalysisTime", props, ""); tmp != "" {
			data = append(data, "total_analysis_time="+tmp)
		}

		if tmp = getStringValueFromMap("afterGreetingSilence", props, ""); tmp != "" {
			data = append(data, "after_greeting_silence="+tmp)
		}

		if tmp = getStringValueFromMap("greeting", props, ""); tmp != "" {
			data = append(data, "greeting="+tmp)
		}

		if tmp = getStringValueFromMap("initialSilence", props, ""); tmp != "" {
			data = append(data, "initial_silence="+tmp)
		}

		_, err := c.SndMsg("amd", strings.Join(data, " "), true, true)
		if err != nil {
			logger.Error("Call %s amd error: ", c.Uuid, err)
			return err
		}
		logger.Debug("Call %s amd successful", c.Uuid)

	} else {
		logger.Error("Call %s AMD bad arguments %s", c.Uuid, args)
	}
	return nil
}
