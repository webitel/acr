/**
 * Created by I. Navrotskyj on 19.08.17.
 */

package acr

import (
	"github.com/webitel/acr/src/pkg/esl"
	"github.com/webitel/acr/src/pkg/logger"
	"github.com/webitel/acr/src/pkg/router"
	"gopkg.in/mgo.v2/bson"
	"strconv"
)

type dialerCallFlowType struct {
	Callflow []interface{}          `bson:"_cf"`
	AMD      map[string]interface{} `bson:"amd"`
}

//originate [^^:domain_name=10.10.10.144:ignore_early_media=true:loopback_bowout=false:dlr_queue=58bd30b0a01699316e2d5ae2]loopback/12312321312312312/default &park()

func dialerContext(a *ACR, c *esl.SConn, destinationNumber, dialerId string) {
	var dialer dialerCallFlowType
	var domainName string
	var answerTime int

	if domainName = c.ChannelData.Header.Get("variable_domain_name"); domainName == "" {
		logger.Error("Call %s no domain", c.Uuid)
		c.SndMsg("hangup", HANGUP_NO_ROUTE_DESTINATION, false, false)
		return
	}

	err := a.DB.FindDialerCallFlow(dialerId, domainName, &dialer)

	if err != nil {
		logger.Error("Call %s find dialer error: %s", c.Uuid, err.Error())
		c.SndMsg("hangup", HANGUP_NO_ROUTE_DESTINATION, false, false)
		return
	}

	if dialer.Callflow == nil {
		logger.Error("Call %s bad dialer call flow: %v", c.Uuid, dialer)
		c.SndMsg("hangup", HANGUP_NORMAL_TEMPORARY_FAILURE, false, false)
		return
	}

	exec := func() {
		cf := router.CallFlow{
			Domain:   domainName,
			Callflow: getDialerRoute(&dialer),
			Version:  2,
		}

		a.CreateCall(destinationNumber, c, &cf)
	}

	answerTime, err = strconv.Atoi(c.ChannelData.Header.Get("Caller-Channel-Answered-Time"))

	if err == nil && answerTime == 0 {
		logger.Debug("Call %s dialer check answer", c.Uuid)
		if c.OnAnswer() {
			exec()
		}
	} else {
		logger.Debug("Call %s dialer answered %s", c.Uuid, c.ChannelData.Header.Get("Caller-Channel-Answered-Time"))
		exec()
	}
	return

}

func getDialerRoute(d *dialerCallFlowType) (r []interface{}) {

	if d.AMD != nil {
		if isTrue("enabled", d.AMD) {
			r = append(r, getAmdExpression(d.AMD)...)
		}
	}

	r = append(r, d.Callflow...)
	r = append(r, bson.M{"hangup": ""})
	return
}

func getAmdExpression(m map[string]interface{}) (r []interface{}) {
	amd := bson.M{}
	ifApp := bson.M{
		"then": []bson.M{
			{
				"hangup": "NORMAL_UNSPECIFIED",
			},
			{
				"break": true,
			},
		},
	}

	var ok bool

	if _, ok = m["maximumWordLength"]; ok {
		amd["maximumWordLength"] = m["maximumWordLength"]
	}
	if _, ok = m["maximumNumberOfWords"]; ok {
		amd["maximumNumberOfWords"] = m["maximumNumberOfWords"]
	}
	if _, ok = m["betweenWordsSilence"]; ok {
		amd["betweenWordsSilence"] = m["betweenWordsSilence"]
	}
	if _, ok = m["minWordLength"]; ok {
		amd["minWordLength"] = m["minWordLength"]
	}
	if _, ok = m["totalAnalysisTime"]; ok {
		amd["totalAnalysisTime"] = m["totalAnalysisTime"]
	}
	if _, ok = m["silenceThreshold"]; ok {
		amd["silenceThreshold"] = m["silenceThreshold"]
	}
	if _, ok = m["afterGreetingSilence"]; ok {
		amd["afterGreetingSilence"] = m["afterGreetingSilence"]
	}
	if _, ok = m["greeting"]; ok {
		amd["greeting"] = m["greeting"]
	}
	if _, ok = m["initialSilence"]; ok {
		amd["initialSilence"] = m["initialSilence"]
	}

	if isTrue("allowNotSure", m) {
		ifApp["expression"] = " !(${amd_result} === 'HUMAN' || ${amd_result} === 'NOTSURE')"
		ifApp["sysExpression"] = "!(sys.getChnVar(\"amd_result\") === 'HUMAN' || sys.getChnVar(\"amd_result\") === 'NOTSURE')"
	} else {
		ifApp["expression"] = "${amd_result} !== 'HUMAN'"
		ifApp["sysExpression"] = "sys.getChnVar(\"amd_result\") !== 'HUMAN'"
	}

	return []interface{}{
		bson.M{"setVar": "ignore_early_media=true"},
		bson.M{"answer": ""},
		bson.M{"amd": amd},
		bson.M{"if": ifApp},
	}
}

func isTrue(name string, m map[string]interface{}) (ok bool) {

	if _, ok = m[name]; ok {
		if _, ok = m[name].(bool); ok {
			return m[name].(bool)
		}
	}
	return
}
