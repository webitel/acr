/**
 * Created by I. Navrotskyj on 30.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/pkg/logger"
	"regexp"
	"strconv"
)

var regValidateJsLocationPattern = regexp.MustCompile(`/?(.*?)/[gimy]*$`)
var regDeleteNumberCharLocationPattern = regexp.MustCompile(`\D`)

type geoLocationT struct {
	Latitude    float64 `bson:"latitude"`
	Longitude   float64 `bson:"longitude"`
	CountryCode string  `bson:"countryCode"`
	Country     string  `bson:"country"`
	City        string  `bson:"city"`
	Type        string  `bson:"type"`
}

func GeoLocation(c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var number, tmp, result string
	var numbers []string
	var i int
	var geoLocation geoLocationT

	if props, ok = args.(map[string]interface{}); !ok {
		logger.Error("Call %s geoLocation bad arguments %s", c.Uuid, args)
		return nil
	}

	if tmp = getStringValueFromMap("variable", props, ""); tmp != "" {
		number = c.GetChannelVar(tmp)
	} else {
		number = c.DestinationNumber
	}

	if number == "" {
		logger.Error("Call %s geoLocation no number %s", c.Uuid, args)
		return nil
	}

	result = getStringValueFromMap("result", props, "")
	tmp = getStringValueFromMap("regex", props, "")

	if result != "" && tmp != "" {
		tmp = regValidateJsLocationPattern.ReplaceAllString(tmp, "$1")
		r := regexp.MustCompile(tmp)
		number = r.ReplaceAllString(number, result)
	}

	number = regDeleteNumberCharLocationPattern.ReplaceAllString(number, "")

	for i, _ = range number {
		numbers = append(numbers, number[:i])
	}

	err := c.acr.FindLocation(len(number), numbers, &geoLocation)
	if err != nil {
		logger.Error("Call %s geoLocation %s db error: %s", c.Uuid, number, err.Error())
		return nil
	}

	if geoLocation.Latitude > 0 && geoLocation.Longitude > 0 {
		numbers = []string{
			"webitel_location=" + strconv.FormatFloat(geoLocation.Latitude, 'G', 5, 32) + ", " +
				strconv.FormatFloat(geoLocation.Longitude, 'G', 5, 32),
			"webitel_location_country=" + geoLocation.Country,
			"webitel_location_type=" + geoLocation.Type,
		}

		if geoLocation.City != "" {
			numbers = append(numbers, "webitel_location_city="+geoLocation.City)
		}

		if geoLocation.CountryCode != "" {
			numbers = append(numbers, "webitel_location_country_code="+geoLocation.CountryCode)
		}

		logger.Debug("Call %s geoLocation set %v", c.Uuid, geoLocation)
		return SetVar(c, numbers)
	} else {
		logger.Warning("Call %s geoLocation bad db response: %v", c.Uuid, geoLocation)
	}

	return nil
}
