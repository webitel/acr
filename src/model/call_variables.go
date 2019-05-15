package model

var MapVariables = map[string]string{
	"state":        "Channel-State",
	"state_number": "Channel-State-Number",
	"channel_name": "Channel-Name",
	"uuid":         "Unique-ID",
	"direction":    "Call-Direction",
	//"state": "Answer-State",  // TODO WTF ??? https://freeswitch.org/confluence/display/FREESWITCH/Channel+Variables
	"read_codec":         "Channel-Read-Codec-Name",
	"read_rate":          "Channel-Read-Codec-Rate",
	"write_codec":        "Channel-Write-Codec-Name",
	"write_rate":         "Channel-Write-Codec-Rate",
	"username":           "Caller-Username",
	"dialplan":           "Caller-Dialplan",
	"caller_id_name":     "Caller-Caller-ID-Name",
	"caller_id_number":   "Caller-Caller-ID-Number",
	"ani":                "Caller-ANI",
	"aniii":              "Caller-Ani-II",
	"network_addr":       "Caller-Network-Addr",
	"destination_number": "Caller-Destination-Number",
	//"uuid": "Caller-Unique-ID", // TODO ??
	"source":  "Caller-Source",
	"context": "Caller-Context",
	"rdnis":   "Caller-Rdnis",
	//"channel_name": "Caller-Channel-Name", // TODO ??
	"profile_index":       "Caller-Profile-Index",
	"created_time":        "Caller-Channel-Created-Time",
	"answered_time":       "Caller-Channel-Answered-Time",
	"hangup_time":         "Caller-Channel-Hangup-Time",
	"transfer_time":       "Caller-Channel-Transfer-Time",
	"screen_bit":          "Caller-Screen-Bit",
	"privacy_hide_name":   "Caller-Privacy-Hide-Name",
	"privacy_hide_number": "Caller-Privacy-Hide-Number",
	"profile_created":     "Caller-Profile-Created-Time",
	// Var
	"sip_h_p-key-flags": "variable_sip_h_P-Key-Flags",
	"sip_h_referred-by": "variable_sip_h_Referred-By",
}
