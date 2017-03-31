/**
 * Created by igor on 27.03.17.
 */

"use strict";

const Call = require(__appRoot + '/router');
    
module.exports = (acr, conn) => {
    new Call(conn, {callflow: test}, acr);
    conn.execute('hangup', "");
};

const test = [
    {
        "log" : "start"
    },
    {
        "tag" : "if",
        "if" : {
            "expression" : "1 == 2",
            "then" : [
                {
                    "log" : "if1"
                }
            ],
            "else" : [
                {
                    "if" : {
                        "expression" : "2 == 2",
                        "then" : [
                            {
                                "log" : "then2"
                            },
                            {
                                "goto" : "if3"
                            },
                            {
                                "if" : {
                                    "expression" : "false",
                                    "then" : [
                                        {
                                            "log" : "then3",
                                            "tag" : "if3"
                                        },
                                        {
                                            "goto" : "else3"
                                        }
                                    ],
                                    "else" : [
                                        {
                                            "log" : "else3",
                                            "tag" : "else3"
                                        },
                                        {
                                            "goto" : "end"
                                        }
                                    ],
                                    "sysExpression" : "false"
                                }
                            }
                        ],
                        "else" : [
                            {
                                "log" : "else2"
                            }
                        ],
                        "sysExpression" : "2 == 2"
                    }
                },
                {
                    "goto" : "def1",
                    "tag" : "else"
                }
            ],
            "sysExpression" : "1 == 2"
        }
    },
    {
        "switch" : {
            "variable" : "${IVR}",
            "case" : {
                "1" : [
                    {
                        "log" : "s1"
                    },
                    {
                        "log" : "6",
                        "tag" : "s1"
                    },
                    {
                        "if" : {
                            "expression" : "true",
                            "then" : [
                                {
                                    "log" : "swIf",
                                    "tag" : "swIf"
                                },
                                {
                                    "goto" : "footer"
                                }
                            ],
                            "else" : [],
                            "sysExpression" : "true"
                        }
                    }
                ],
                "default" : [
                    {
                        "log" : "5",
                        "tag" : "def1"
                    },
                    {
                        "goto" : "s1"
                    }
                ]
            }
        }
    },
    {
        "tag" : "end",
        "goto" : "s1"
    },
    {
        "tag" : "footer",
        "log" : "footer"
    }
];

const test2 = [
    {
        "tag": "fake",
        "if": {
            "expression": "1 === 1",
            "then": [
                {
                    "log": "ANSWER"
                },
                {
                    "if": {
                        "expression": "2 === 2",
                        "then": [
                            {
                                "log": "2 === 2 : OK",
                                "tag": "fake2"
                            },
                            {
                                "if": {
                                    "expression": "3 === 3",
                                    "then": [
                                        {
                                            "log": "3 === 3 : OK"
                                        },
                                        {
                                            "fn": "asd"
                                        },
                                        {
                                            "goto": "fake"
                                        }
                                    ],
                                    "else": [
                                        {
                                            "log": "3 === 3 : ERR"
                                        }
                                    ]
                                }
                            }
                        ],
                        "else": [
                            {
                                "log": "2 === 2 : ERR"
                            }
                        ]
                    }
                }
            ]
        }
    },
    {
        "log": "sleep"
    },
    {
        "tag": "end",
        "log": "END!"
    }
]