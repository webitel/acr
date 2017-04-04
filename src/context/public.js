/**
 * Created by igor on 27.03.17.
 */

"use strict";

const Call = require(__appRoot + '/router');
    
module.exports = (acr, conn) => {
    // new Call(conn, {callflow: test}, acr);

    const setSound = () => {
        conn.execute('set', "sound_prefix=/$${sounds_dir}/en/us/callie", setTime);
    };


    const setTime = (cb) => {
        conn.execute('set', "timezone=Europe/Kiev", setDomain);
    };

    const setDomain = (cb) => {
        conn.execute('set', "domain_name=10.10.10.144", setTransfer);
    };

    const setTransfer = (cb) => {
        conn.execute('set', "force_transfer_context=default", setHash);
    };

    const setHash = (cb) => {
        setMulti();
        //conn.execute('hash', "insert/10.10.10.144-last_dial/global/${uuid}", setMulti);
    };

    const setMulti = (cb) => {
        setSleep()
        // conn.execute('multiset', "^^~eavesdrop_group=10.10.10.144~presence_data=10.10.10.144", setSleep);
    };

    const setSleep = (cb) => {
        end();
        // conn.execute('sleep', "5000", end);
    };

    const end = (cb) => {
        conn.execute('answer', "", e => {
            // console.log('END');
        });
    };


    setSound()

};

const test = [
    {
        "function" : {
            "name" : "testFn",
            "actions" : [
                {
                    "abstract" : "ok my function"
                },
                {
                    "if" : {
                        "expression" : "1===1",
                        "then" : [
                            {
                                "abstract" : "fn true",
                                "tag" : "isolate"
                            },
                            {
                                "goto" : "else",
                                "tag" : "then"
                            }
                        ],
                        "else" : [
                            {
                                "tag" : "else",
                                "abstract" : "elseFn"
                            },
                            {
                                "goto" : "then"
                            }
                        ],
                        "sysExpression" : "1===1"
                    }
                }
            ]
        }
    },
    {
        "execute" : "testFn",
        "tag" : "else"
    },
    {
        "execute" : "testFn"
    },
    {
        "goto" : "isolate"
    },
    {
        "abstract" : "start"
    },
    {
        "tag" : "if",
        "if" : {
            "expression" : "1 == 2",
            "then" : [
                {
                    "abstract" : "if1"
                }
            ],
            "else" : [
                {
                    "if" : {
                        "expression" : "2 == 2",
                        "then" : [
                            {
                                "abstract" : "then2"
                            },
                            {
                                "goto" : "if3"
                            },
                            {
                                "if" : {
                                    "expression" : "false",
                                    "then" : [
                                        {
                                            "abstract" : "then3",
                                            "tag" : "if3"
                                        },
                                        {
                                            "goto" : "else3"
                                        }
                                    ],
                                    "else" : [
                                        {
                                            "abstract" : "else3",
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
                                "abstract" : "else2"
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
                        "abstract" : "s1"
                    },
                    {
                        "abstract" : "6",
                        "tag" : "s1"
                    },
                    {
                        "if" : {
                            "expression" : "true",
                            "then" : [
                                {
                                    "abstract" : "swIf",
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
                        "abstract" : "5",
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
        "abstract" : "enad call ${uuid}"
    }
];
