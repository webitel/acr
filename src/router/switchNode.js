/**
 * Created by igor on 29.03.17.
 */

"use strict";
    
const BaseNode = require('./baseNode'),
    log = require(__appRoot + '/lib/log')(module),
    Node = require('./node');

/*
{
    "switch": {
        "variable": "${IVR}",
        "case": {
            "1": [],
            "2": [],
            "3": [],
            "default": []
        }
    }
}
*/


class SwitchNode extends BaseNode {
    constructor (parent, args = {}, options = {}) {
        super(parent, options);

        this.nameVariable = args.variable;
        this.case = new Map();
        this._values = [];

        this.getCaseWorkFlow = (value) => {
            return args.case[value];
        };

        if (args.case instanceof Object) {
            this._values = Object.keys(args.case);
            this._values.forEach( value => {
                this.case.set(value, new Node(parent))
            });
        }
    }

    getValueNode (value) {
        const node = this.case.get(value);
        node.first();
        return node;
    }

    getValues () {
        return this._values;
    }

    execute (call, cb) {
        const value = null; //TODO

        for (let val of this._values) {
            if (value == val) {
                call.callFlowIter.setRoot(this.getValueNode(val));
                log.trace(`Switch value ${value} execute case`);
                return cb();
            }
        }

        if (this.case.has('default')) {
            call.callFlowIter.setRoot(this.getValueNode('default'));
            log.trace(`Switch value ${value} not found case, execute default`);
            return cb();
        }

        log.trace(`Switch value ${value} not found route`);
        return cb();
    }
}

module.exports = SwitchNode;