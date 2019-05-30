package model

const (
	CONTEXT_DEFAULT = "default"
	CONTEXT_PUBLIC  = "public"
	CONTEXT_DIALER  = "dialer"
	CONTEXT_PRIVATE = "private"
)

const (
	CALL_DIRECTION_INBOUND  = "inbound"
	CALL_DIRECTION_OUTBOUND = "outbound"
	CALL_DIRECTION_INTERNAL = "internal"
)

const (
	CALL_BRIDGE_USER_TEMPLATE = "sofia/sip/%s@%s"
)

const (
	CALL_VARIABLE_TIMEZONE_NAME          = "timezone"
	CALL_VARIABLE_DOMAIN_NAME            = "sip_h_X-Webitel-Domain"
	CALL_VARIABLE_DIRECTION_NAME         = "sip_h_X-Webitel-Direction"
	CALL_VARIABLE_FORCE_TRANSFER_CONTEXT = "force_transfer_context"

	CALL_VARIABLE_SHEMA_ID   = "webitel_acr_schema_id"
	CALL_VARIABLE_SHEMA_NAME = "webitel_acr_schema_name"
	CALL_VARIABLE_DEBUG_NAME = "webitel_debug_acr"
	CALL_VARIABLE_DIALER_ID  = "variable_dlr_queue"

	CALL_VARIABLE_DEFAULT_LANGUAGE_NAME = "default_language"
	CALL_VARIABLE_SOUND_PREF_NAME       = "sound_prefix"
)

const (
	GLOBAL_VARIABLE_DEFAULT_PUBLIC_NAME = "webitel_default_public_route"
)

const (
	CALL_LANGUAGE_RU                = "ru"
	CALL_LANGUAGE_RU_DIRECTORY      = "/$${sounds_dir}/ru/RU/elena"
	CALL_LANGUAGE_DEFAULT_DIRECTORY = "/$${sounds_dir}/en/us/callie"
)

const (
	HANGUP_NORMAL_TEMPORARY_FAILURE = "NORMAL_TEMPORARY_FAILURE"
	HANGUP_NO_ROUTE_DESTINATION     = "NO_ROUTE_DESTINATION"
)
