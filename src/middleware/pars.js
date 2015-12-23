/**
 * Created by i.navrotskyj on 14.12.2015.
 */
'use strict';

const OPERATION = {
    IF: "if",
    THEN: "then",
    ELSE: "else",
    SWITCH: "switch",
    APPLICATION: "app",
    DATA: "data",
    ASYNC: "async",

    ECHO: "echo",

    ANSWER: "answer",
    SET: "setVar",
    GOTO: "goto",
    /* GATEWAY: "gateway",
     DEVICE: "device", */
    RECORD_SESSION: "recordSession",
    RECORD_FILE: "recordFile",
    HANGUP: "hangup",
    SCRIPT: "script",
    LOG: "log",
    HTTP: "httpRequest",
    SLEEP: "sleep",

    CONFERENCE: "conference",

    SCHEDULE: "schedule",

    BRIDGE: "bridge",
    PLAYBACK: "playback",

    BREAK: "break",

    CALENDAR: "calendar",
    PARK: "park",
    QUEUE: "queue",
    CC_POSITION: "ccPosition",

    EXPORT_VARS: "exportVars",
    VOICEMAIL: "voicemail",

    IVR: "ivr",

    BIND_ACTION: "bindAction",
    CLEAR_ACTION: "clearAction",

    BIND_EXTENSION: "bindExtension",

    ATT_XFER: 'attXfer',

    UN_SET: 'unSet',

    SET_USER: 'setUser',
    CALL_FORWARD: 'checkCallForward',
    RECEIVE_FAX: 'receiveFax',

    TAGS: 'setArray',
    BLACK_LIST: 'blackList',

    PICKUP: 'pickup',

    DISA: 'disa',

    SEND_SMS: 'sendSms',

    LOCATION: "geoLocation",
    RINGBACK: "ringback",

    SET_SOUNDS: "setSounds",
    EVENT: "event",

    IN_BAND_DTMF: 'inBandDTMF',
    FLUSH_DTMF: 'flushDTMF',
    EMAIL: 'sendEmail'
};

const _APP_ARRAY = (function () {
    var a = [];
    for (var key in OPERATION) {
        if (OPERATION.hasOwnProperty(key))
            a.push(OPERATION[key])
    };
    return a;
})();

function isApplication (obj) {
    if (obj instanceof Object) {
        for (let key in obj) {
            if (obj.hasOwnProperty(key) && _APP_ARRAY.indexOf(key) != -1)
                return key
        }
    };
    return false;
};

String.prototype.hashCode = function() {
    var hash = 0, i, chr, len;
    if (this.length === 0) return hash;
    for (i = 0, len = this.length; i < len; i++) {
        chr   = this.charCodeAt(i);
        hash  = ((hash << 5) - hash) + chr;
        hash |= 0; // Convert to 32bit integer
    }
    return hash;
};

var cf = [
    {
        "geoLocation": {
            "variable": "Caller-Caller-ID-Number",
            "regex": "^(\\d{10})",
            "result": "38$1"
        }
    },
    {
        "ringback": {
            "call": {
                "name": "$${ru-ring}",
                "type": "tone"
            },
            "transfer": {
                "name": "$${ru-ring}",
                "type": "tone"
            }
        }
    },
    {
        "answer": "200"
    },
    {
        "recordSession": {
            "action": "start",
            "stereo": "false"
        }
    },
    {
        "playback": {
            "name": "1m.mp3",
            "type": "mp3",
            "getDigits": {
                "setVar": "getIvrLang",
                "min": "1",
                "max": "1",
                "tries": "2",
                "timeout": "5000"
            }
        },
        "tag": "IvrLang"
    },
    {
        "switch": {
            "variable": "${getIvrLang}",
            "case": {
                "1": [
                    {
                        "setVar": "ivrLang=ua"
                    }
                ],
                "2": [
                    {
                        "setVar": "ivrLang=ru"
                    }
                ],
                "3": [
                    {
                        "setVar": "ivrLang=en"
                    },
                    {
                        "goto": "local:Operator"
                    }
                ],
                "default": [
                    {
                        "goto": "local:IvrLang"
                    }
                ]
            }
        }
    },
    {
        "flushDTMF": true
    },
    {
        "playback": {
            "name": "2m_${ivrLang}.mp3",
            "type": "mp3",
            "getDigits": {
                "setVar": "getMainMenuAction",
                "min": "1",
                "max": "1",
                "tries": "2",
                "timeout": "5000"
            }
        },
        "tag": "MainMenu"
    },
    {
        "switch": {
            "variable": "${getMainMenuAction}",
            "case": {
                "1": [
                    {
                        "setVar": "mainMenuAction=Блокування карти"
                    },
                    {
                        "goto": "local:Operator"
                    }
                ],
                "2": [
                    {
                        "setVar": "mainMenuAction=Обсуговування платіжних карток"
                    },
                    {
                        "goto": "local:Cards"
                    }
                ],
                "3": [
                    {
                        "setVar": "mainMenuAction=Обсуговування депозитних продуктів"
                    },
                    {
                        "playback": {
                            "name": "4m_${ivrLang}.mp3",
                            "type": "mp3"
                        }
                    },
                    {
                        "if": {
                            "expression": "&wday(2-6) && &time_of_day(08:35-13:00, 14:00-17:30)",
                            "then": [
                                {
                                    "setVar": [
                                        "continue_on_fail=true",
                                        "hangup_after_bridge=true"
                                    ]
                                },
                                {
                                    "bridge": {
                                        "endpoints": [
                                            {
                                                "type": "sipGateway",
                                                "dialString": "1151",
                                                "name": "sbc-d1",
                                                "parameters": [
                                                    "leg_timeout=30"
                                                ]
                                            },
                                            {
                                                "type": "sipGateway",
                                                "dialString": "1152",
                                                "name": "sbc-d1",
                                                "parameters": [
                                                    "leg_timeout=30"
                                                ]
                                            }
                                        ],
                                        "parameters": [
                                            "absolute_codec_string='PCMA'",
                                            "instant_ringback=true"
                                        ]
                                    }
                                }
                            ],
                            "sysExpression": "sys.wday(\"2-6\") && sys.time_of_day(\"08:35-13:00, 14:00-17:30\")"
                        }
                    },
                    {
                        "goto": "local:Consultant"
                    }
                ],
                "4": [
                    {
                        "setVar": "mainMenuAction=Обсуговування кредитних продуктів"
                    },
                    {
                        "goto": "local:Credit"
                    }
                ],
                "5": [
                    {
                        "setVar": "mainMenuAction=Обсуговування юридичних осіб"
                    },
                    {
                        "playback": {
                            "name": "4m_${ivrLang}.mp3",
                            "type": "mp3"
                        }
                    },
                    {
                        "if": {
                            "expression": "&wday(2-6) && &time_of_day(08:35-13:00, 14:00-17:30)",
                            "then": [
                                {
                                    "setVar": [
                                        "continue_on_fail=true",
                                        "hangup_after_bridge=true"
                                    ]
                                },
                                {
                                    "bridge": {
                                        "endpoints": [
                                            {
                                                "type": "sipGateway",
                                                "dialString": "1191",
                                                "name": "sbc-d1",
                                                "parameters": [
                                                    "leg_timeout=30"
                                                ]
                                            },
                                            {
                                                "type": "sipGateway",
                                                "dialString": "1188",
                                                "name": "sbc-d1",
                                                "parameters": [
                                                    "leg_timeout=30"
                                                ]
                                            }
                                        ],
                                        "parameters": [
                                            "absolute_codec_string='PCMA'",
                                            "instant_ringback=true"
                                        ]
                                    }
                                }
                            ],
                            "sysExpression": "sys.wday(\"2-6\") && sys.time_of_day(\"08:35-13:00, 14:00-17:30\")"
                        }
                    },
                    {
                        "goto": "local:Consultant"
                    }
                ],
                "6": [
                    {
                        "setVar": "mainMenuAction=Інформація по комунальних платежах"
                    },
                    {
                        "playback": {
                            "name": "4m_${ivrLang}.mp3",
                            "type": "mp3"
                        }
                    },
                    {
                        "if": {
                            "expression": "&wday(2-6) && &time_of_day(08:35-13:00, 14:00-17:30)",
                            "then": [
                                {
                                    "setVar": [
                                        "continue_on_fail=true",
                                        "hangup_after_bridge=true"
                                    ]
                                },
                                {
                                    "bridge": {
                                        "endpoints": [
                                            {
                                                "type": "sipGateway",
                                                "dialString": "1127",
                                                "name": "sbc-d1",
                                                "parameters": [
                                                    "leg_timeout=30"
                                                ]
                                            },
                                            {
                                                "type": "sipGateway",
                                                "dialString": "1128",
                                                "name": "sbc-d1",
                                                "parameters": [
                                                    "leg_timeout=30"
                                                ]
                                            }
                                        ],
                                        "parameters": [
                                            "absolute_codec_string='PCMA'",
                                            "instant_ringback=true"
                                        ]
                                    }
                                }
                            ],
                            "sysExpression": "sys.wday(\"2-6\") && sys.time_of_day(\"08:35-13:00, 14:00-17:30\")"
                        }
                    },
                    {
                        "goto": "local:Consultant"
                    }
                ],
                "7": [
                    {
                        "setVar": "mainMenuAction=Надати пропозиції"
                    },
                    {
                        "playback": {
                            "name": "5m_${ivrLang}.mp3",
                            "type": "mp3"
                        }
                    },
                    {
                        "playback": {
                            "name": "%(1000, 0, 640)",
                            "type": "tone"
                        }
                    },
                    {
                        "recordFile": {
                            "name": "ContactCenterSuggestions",
                            "type": "mp3",
                            "maxSec": "60",
                            "email": [
                                "ContactCenterSuggestions@megabank.net"
                            ]
                        }
                    },
                    {
                        "flushDTMF": true
                    },
                    {
                        "playback": {
                            "name": "10m_${ivrLang}.mp3",
                            "type": "mp3",
                            "getDigits": {
                                "setVar": "getSubMenuAction",
                                "min": "1",
                                "max": "1",
                                "tries": "2",
                                "timeout": "5000"
                            }
                        }
                    },
                    {
                        "if": {
                            "expression": "${getSubMenuAction} == '6'",
                            "then": [
                                {
                                    "goto": "local:MainMenu"
                                }
                            ],
                            "else": [
                                {
                                    "playback": {
                                        "name": "L=10;%(400,400,425)",
                                        "type": "tone"
                                    }
                                },
                                {
                                    "hangup": "",
                                    "break": true
                                }
                            ],
                            "sysExpression": "sys.getChnVar(\"getSubMenuAction\") == '6'"
                        }
                    }
                ],
                "default": [
                    {
                        "setVar": "mainMenuAction=На оператора"
                    },
                    {
                        "goto": "local:Operator"
                    }
                ]
            }
        }
    },
    {
        "log": "[MEGABANK]: ${mainMenuAction}",
        "tag": "Cards"
    },
    {
        "flushDTMF": true
    },
    {
        "playback": {
            "name": "6m_${ivrLang}.mp3",
            "type": "mp3",
            "getDigits": {
                "setVar": "getSubMenuAction",
                "min": "1",
                "max": "1",
                "tries": "3",
                "timeout": "5000"
            }
        }
    },
    {
        "switch": {
            "variable": "${getSubMenuAction}",
            "case": {
                "1": [
                    {
                        "setVar": "subMenuAction=Баланс по карті"
                    },
                    {
                        "goto": "local:Operator"
                    }
                ],
                "2": [
                    {
                        "setVar": "subMenuAction=Зняття платіжних лімітів"
                    },
                    {
                        "goto": "local:Operator"
                    }
                ],
                "3": [
                    {
                        "setVar": "subMenuAction=Заборгованість за овердрафтом"
                    },
                    {
                        "goto": "local:Operator"
                    }
                ],
                "4": [
                    {
                        "setVar": "subMenuAction=Встановити SMS інформування"
                    },
                    {
                        "goto": "local:CardToEmail"
                    }
                ],
                "5": [
                    {
                        "setVar": "subMenuAction=Продовжити термін дії карти"
                    },
                    {
                        "goto": "local:CardToEmail"
                    }
                ],
                "default": [
                    {
                        "setVar": "subMenuAction=В головне меню"
                    },
                    {
                        "goto": "local:MainMenu"
                    }
                ]
            }
        }
    },
    {
        "log": "[MEGABANK]: ${mainMenuAction}",
        "tag": "CardToEmail"
    },
    {
        "flushDTMF": true
    },
    {
        "playback": {
            "name": "7m_1_${ivrLang}.mp3",
            "type": "mp3",
            "getDigits": {
                "setVar": "getSubMenuAction",
                "min": "6",
                "max": "6",
                "tries": "10",
                "timeout": "10000"
            }
        }
    },
    {
        "setArray": {
            "Creditcard": [
                "${getSubMenuAction}"
            ]
        }
    },
    {
        "flushDTMF": true
    },
    {
        "sleep": "2000"
    },
    {
        "playback": {
            "name": "7m_2_${ivrLang}.mp3",
            "type": "mp3",
            "getDigits": {
                "setVar": "getSubMenuAction",
                "min": "4",
                "max": "4",
                "tries": "10",
                "timeout": "10000"
            }
        }
    },
    {
        "setArray": {
            "Creditcard": [
                "${getSubMenuAction}"
            ]
        }
    },
    {
        "sleep": "2000"
    },
    {
        "flushDTMF": true
    },
    {
        "playback": {
            "name": "7m_3_${ivrLang}.mp3",
            "type": "mp3",
            "getDigits": {
                "setVar": "getSubMenuAction",
                "min": "4",
                "max": "4",
                "tries": "10",
                "timeout": "10000"
            }
        }
    },
    {
        "setArray": {
            "Creditcard": [
                "${getSubMenuAction}"
            ]
        }
    },
    {
        "flushDTMF": true
    },
    {
        "sleep": "2000"
    },
    {
        "playback": {
            "name": "7m_4_${ivrLang}.mp3",
            "type": "mp3",
            "getDigits": {
                "setVar": "getSubMenuAction",
                "min": "6",
                "max": "8",
                "tries": "10",
                "timeout": "8000"
            }
        }
    },
    {
        "setArray": {
            "Creditcard": [
                "${getSubMenuAction}"
            ]
        }
    },
    {
        "flushDTMF": true
    },
    {
        "if": {
            "expression": "${subMenuAction} == 'Встановити SMS інформування'",
            "then": [
                {
                    "sendEmail": {
                        "to": [
                            "ContactCenterSMSInform@megabank.net "
                        ],
                        "subject": "[webitel](${Caller-Caller-ID-Number}) Встановить SMS інформування",
                        "message": "<H2>Встановити SMS інформування (${Caller-Caller-ID-Number})</h2>\n<b>Номер карти</b>: ${Creditcard[0]} <i>***</i> ${Creditcard[1]}<br />\n<b>Дійсна до</b>: ${Creditcard[2]}<br />\n<b>Дата народження</b>: ${Creditcard[3]}"
                    }
                },
                {
                    "playback": {
                        "name": "8m_${ivrLang}.mp3",
                        "type": "mp3"
                    }
                }
            ],
            "else": [
                {
                    "sendEmail": {
                        "to": [
                            "ContactCenterProlongCard@megabank.net"
                        ],
                        "subject": "[webitel](${Caller-Caller-ID-Number}) Продовжити термін дії карти",
                        "message": "<H2>Продовжити термін дії карти (${Caller-Caller-ID-Number})</h2>\n<b>Номер карти</b>: ${Creditcard[0]} <i>***</i> ${Creditcard[1]}<br />\n<b>Дійсна до</b>: ${Creditcard[2]}<br />\n<b>Дата народження</b>: ${Creditcard[3]}"
                    }
                },
                {
                    "playback": {
                        "name": "9m_${ivrLang}.mp3",
                        "type": "mp3"
                    }
                }
            ],
            "sysExpression": "sys.getChnVar(\"subMenuAction\") == 'Встановити SMS інформування'"
        }
    },
    {
        "goto": "local:MainMenu"
    },
    {
        "log": "[MEGABANK]: ${mainMenuAction}",
        "tag": "Credit"
    },
    {
        "flushDTMF": true
    },
    {
        "playback": {
            "name": "11m_${ivrLang}.mp3",
            "type": "mp3",
            "getDigits": {
                "setVar": "getSubMenuAction",
                "min": "1",
                "max": "1",
                "tries": "3",
                "timeout": "5000"
            }
        },
        "tag": "SubMenuAction2"
    },
    {
        "switch": {
            "variable": "${getSubMenuAction}",
            "case": {
                "1": [
                    {
                        "setVar": "subMenuAction=Експрес-кредит"
                    },
                    {
                        "goto": "local:SubMenuAction3"
                    }
                ],
                "2": [
                    {
                        "setVar": "subMenuAction=По іншому кредитному продукту"
                    },
                    {
                        "if": {
                            "expression": "&wday(2-6) && &time_of_day(08:35-13:00, 14:00-17:30)",
                            "then": [
                                {
                                    "setVar": [
                                        "continue_on_fail=true",
                                        "hangup_after_bridge=true"
                                    ]
                                },
                                {
                                    "bridge": {
                                        "endpoints": [
                                            {
                                                "type": "sipGateway",
                                                "dialString": "1234",
                                                "name": "sbc-d1",
                                                "parameters": [
                                                    "leg_timeout=30"
                                                ]
                                            },
                                            {
                                                "type": "sipGateway",
                                                "dialString": "1205",
                                                "name": "sbc-d1",
                                                "parameters": [
                                                    "leg_timeout=30"
                                                ]
                                            }
                                        ],
                                        "parameters": [
                                            "absolute_codec_string='PCMA'",
                                            "instant_ringback=true"
                                        ]
                                    }
                                }
                            ],
                            "sysExpression": "sys.wday(\"2-6\") && sys.time_of_day(\"08:35-13:00, 14:00-17:30\")"
                        }
                    },
                    {
                        "goto": "local:Consultant"
                    }
                ],
                "3": [
                    {
                        "setVar": "subMenuAction=Сума просроченої заборгованості"
                    },
                    {
                        "if": {
                            "expression": "&wday(2-6) && &time_of_day(08:35-13:00, 14:00-17:30)",
                            "then": [
                                {
                                    "setVar": [
                                        "continue_on_fail=true",
                                        "hangup_after_bridge=true"
                                    ]
                                },
                                {
                                    "bridge": {
                                        "endpoints": [
                                            {
                                                "type": "sipGateway",
                                                "dialString": "1234",
                                                "name": "sbc-d1",
                                                "parameters": [
                                                    "leg_timeout=30"
                                                ]
                                            },
                                            {
                                                "type": "sipGateway",
                                                "dialString": "1205",
                                                "name": "sbc-d1",
                                                "parameters": [
                                                    "leg_timeout=30"
                                                ]
                                            }
                                        ],
                                        "parameters": [
                                            "absolute_codec_string='PCMA'",
                                            "instant_ringback=true"
                                        ]
                                    }
                                }
                            ],
                            "sysExpression": "sys.wday(\"2-6\") && sys.time_of_day(\"08:35-13:00, 14:00-17:30\")"
                        }
                    },
                    {
                        "goto": "local:Consultant"
                    }
                ],
                "default": [
                    {
                        "setVar": "subMenuAction=В головне меню"
                    },
                    {
                        "goto": "local:MainMenu"
                    }
                ]
            }
        }
    },
    {
        "log": "[MEGABANK]: ${mainMenuAction}",
        "tag": "Operator"
    },
    {
        "exportVars": [
            "ivrLang",
            "mainMenuAction",
            "subMenuAction"
        ]
    },
    {
        "playback": {
            "name": "3m_${ivrLang}.mp3",
            "type": "mp3"
        }
    },
    {
        "queue": {
            "name": "${ivrLang}"
        },
        "break": true
    },
    {
        "flushDTMF": true
    },
    {
        "playback": {
            "name": "12m_${ivrLang}.mp3",
            "type": "mp3",
            "getDigits": {
                "setVar": "getSubMenuAction",
                "min": "1",
                "max": "1",
                "tries": "3",
                "timeout": "5000"
            }
        },
        "tag": "SubMenuAction3"
    },
    {
        "switch": {
            "variable": "${getSubMenuAction}",
            "case": {
                "2": [
                    {
                        "goto": "local:SubMenuAction3"
                    }
                ],
                "3": [
                    {
                        "setVar": "subMenuAction=Оформити експрес-кредит"
                    },
                    {
                        "if": {
                            "expression": "&wday(2-6) && &time_of_day(08:35-13:00, 14:00-17:30)",
                            "then": [
                                {
                                    "setVar": [
                                        "continue_on_fail=true",
                                        "hangup_after_bridge=true"
                                    ]
                                },
                                {
                                    "bridge": {
                                        "endpoints": [
                                            {
                                                "type": "sipGateway",
                                                "dialString": "1234",
                                                "name": "sbc-d1",
                                                "parameters": [
                                                    "leg_timeout=30"
                                                ]
                                            },
                                            {
                                                "type": "sipGateway",
                                                "dialString": "1205",
                                                "name": "sbc-d1",
                                                "parameters": [
                                                    "leg_timeout=30"
                                                ]
                                            }
                                        ],
                                        "parameters": [
                                            "absolute_codec_string='PCMA'",
                                            "instant_ringback=true"
                                        ]
                                    }
                                }
                            ],
                            "sysExpression": "sys.wday(\"2-6\") && sys.time_of_day(\"08:35-13:00, 14:00-17:30\")"
                        }
                    }
                ],
                "default": [
                    {
                        "setVar": "subMenuAction=В головне меню"
                    },
                    {
                        "goto": "local:MainMenu"
                    }
                ]
            }
        }
    },
    {
        "log": "[MEGABANK]: ${mainMenuAction}",
        "tag": "Consultant"
    },
    {
        "exportVars": [
            "ivrLang",
            "mainMenuAction",
            "subMenuAction"
        ]
    },
    {
        "queue": {
            "name": "${ivrLang}"
        },
        "break": true
    }
];

var log = console

function getStatesRoute(data) {
    var validId = /^[a-z_$][a-z0-9_$]*$/i;
    var result = [];
    doIt(data, "");

    var nodes = [];
    var edges = [];
    for (let i in result) {
        console.log('NAME:  ', result[i].name, '     STATE:', result[i].state,'    NEXT ->>>', result[i].next);

        nodes.push({
            "id": result[i].state.hashCode(),
            "_hash": result[i].state,
            "label": result[i].name + '(' + result[i].state +')'
        });
        edges.push({
            "from": result[i].state.hashCode(),
            "to": (result[i].next || '').hashCode()
        });
    };

    console.log(JSON.stringify(nodes));
    console.log(JSON.stringify(edges));
    return result;

    function doIt(data, s, prev) {
        if (data && typeof data === "object") {
            if (Array.isArray(data)) {
                for (var i = 0; i < data.length; i++) {
                    var app = isApplication(data[i]);
                    if (app) {
                        var next = (data[i + 1])
                            ? s + "[" + (i + 1) + "]"
                            : prev;

                        result.push({
                            "state": s + "[" + i + "]",
                            "next": next,
                            "name": app,
                            "application": data[i]
                        });
                    } else {
                       // log.warn('Skip application', data[i]);
                    };
                    doIt(data[i], s + "[" + i + "]", app == 'if' || app == 'switch'  ? s + "[" + (i + 1) + "]" : prev);
                }
            } else {
                for (var p in data) {
                    if (validId.test(p)) {
                        doIt(data[p], s + "." + p, prev);
                    } else {
                        doIt(data[p], s + "[\"" + p + "\"]", prev);
                    }
                }
            }
        }
    }
};

var b = getStatesRoute(cf);
console.log(b);






class Executor {
    constructor (cf) {
        this.index = 0;
        this._cf = cf || [];
        this.pointer = '';
        this.setHash(this._cf);
    };

    setHash (data) {
        var validId = /^[a-z_$][a-z0-9_$]*$/i;
        doIt(data, "");
        var tree = {};
        return tree;

        function doIt(data, s) {
            if (data && typeof data === "object") {
                if (Array.isArray(data)) {
                    for (var i = 0; i < data.length; i++) {
                        doIt(data[i], s + "[" + i + "]");
                    }
                } else {
                    data['_hash'] = s;
                    for (var p in data) {
                        if (validId.test(p)) {
                            doIt(data[p], s + "." + p);
                        } else {
                            doIt(data[p], s + "[\"" + p + "\"]");
                        }
                    }
                }
            }
        }
    };

    exec (app, cb) {
        console.log(app);

    };

    setState (err, res, state) { // state == string
        var current = this._cf;
        ('' + state).split('.').forEach(function(token) {
            current = current && current[token];
        });
        this.exec(current, this.setState);
    };

    run() {
        this.setState(null, null, 0);
    }

};


//
//var a = new Executor(cf);
//a.run();


var SM = function(a, b) {
    return {
        event: function(c) {
            return (c = b[a][c]) && ((c[0][0] || c[0]).apply(c[0][1], [].slice.call(arguments, 1)), a = c[1] || a)
        }
    }
};




/*

function objectToPaths(data) {
    var validId = /^[a-z_$][a-z0-9_$]*$/i;
    var result = {};
    doIt(data, "");
    return result;

    function doIt(data, s, _k) {
        if (data && typeof data === "object") {
            if (Array.isArray(data)) {
                for (var i = 0; i < data.length; i++) {
                    doIt(data[i], s + "[" + i + "]", i);
                }
            } else {
                for (var p in data) {
                    if (validId.test(p)) {
                        doIt(data[p], s + "." + p, p);
                    } else {
                        doIt(data[p], s + "[\"" + p + "\"]", p);
                    }
                }
            }
        } else if (_k == 'tag' && typeof data == 'string' && !result[data]) {
            result[data] = s.replace(/.tag$/, '');
        }
    }
};


function addHashPath(data) {
    var validId = /^[a-z_$][a-z0-9_$]*$/i;
    doIt(data, "");
    return data;

    function doIt(data, s) {
        if (data && typeof data === "object") {
            if (Array.isArray(data)) {
                for (var i = 0; i < data.length; i++) {
                    doIt(data[i], s + "[" + i + "]");
                }
            } else {
                data['_hash'] = s;

                for (var p in data) {
                    if (validId.test(p)) {
                        doIt(data[p], s + "." + p);
                    } else {
                        doIt(data[p], s + "[\"" + p + "\"]");
                    }
                }
            }
        }
    }
}

console.time('Parse')
var d = addHashPath(cf);
console.log(d)
console.timeEnd('Parse')

*/