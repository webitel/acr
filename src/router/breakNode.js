/**
 * Created by igor on 29.03.17.
 */

"use strict";

const BaseNode = require('./baseNode');
    
class BreakNode extends BaseNode {
    constructor(parent) {
        super(parent);
        this.break = true;
    }

    execute (call, cb) {
        return cb();
    }
}

module.exports = BreakNode;