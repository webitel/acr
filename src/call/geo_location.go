/**
 * Created by I. Navrotskyj on 30.08.17.
 */

package call

import (
	"github.com/webitel/acr/src/model"
	"regexp"
	"strconv"
)

var regValidateJsLocationPattern = regexp.MustCompile(`/?(.*?)/[gimy]*$`)
var regDeleteNumberCharLocationPattern = regexp.MustCompile(`\D`)

func GeoLocation(scope Scope, c *Call, args interface{}) error {
	var props map[string]interface{}
	var ok bool
	var number, tmp, result string
	var numbers []string
	var i int
	var geoLocation *model.GeoLocation

	if props, ok = args.(map[string]interface{}); !ok {
		c.LogError("geoLocation", args, "bad request")
		return nil
	}

	if tmp = getStringValueFromMap("variable", props, ""); tmp != "" {
		number = c.GetVariable(tmp)
	} else {
		number = c.Destination()
	}

	if number == "" {
		c.LogError("geoLocation", props, "number is required")
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

	res := <-c.router.app.Store.GeoLocation().Find(len(number), numbers)
	if res.Err != nil {
		c.LogError("geoLocation", props, res.Err.Error())
		return nil
	}

	geoLocation = res.Data.(*model.GeoLocation)

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

		c.LogDebug("geoLocation", geoLocation, "success")
		return SetVar(scope, c, numbers)
	} else {
		c.LogWarn("geoLocation", props, "not found")
	}

	return nil
}
