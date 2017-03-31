/**
 * Created by igor on 29.03.17.
 */

"use strict";

const cf = [
    {
        "tag": "if",
        "if": {
            "expression": "1 == 1",
            "then": [
                {
                    "log": "then1",
                    "tag": "yo",
                    "async": true,
                    "dump": true
                },
                {
                    "log": "then2",
                    "tag": "yo",
                    "async": true,
                    "dump": true
                },
                {
                    "log": "fack of",
                    "@break": true,
                    "tag": "fake"
                },
                {
                    "log": "then3",
                    "tag": "yo",
                    "async": true,
                    "dump": true
                },
                {
                    "goto": "else"
                }
            ],
            "else": [
                {
                    "goto": "def1",
                    "tag": "else"
                }
            ]
        }
    },
    {
        "switch": {
            "variable": "${IVR}",
            "case": {
                "1": [
                    {
                        "log": "s1"
                    },
                    {
                        "log": "s11",
                        "tag": "s1"
                    }
                ],
                "default": [
                    {
                        "log": "def1",
                        "tag": "def1"
                    },
                    {
                        "goto": "s1"
                    }
                ]
            }
        }
    },
    {
        "tag": "end",
        "goto": "fake"
    },
    {
        "log": "footer"
    }
]