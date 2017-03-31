/**
 * Created by igor on 29.03.17.
 */

"use strict";

const BaseNode = require('./baseNode');

class ApplicationNode extends BaseNode {
    constructor (parent, name, args = {}, options = {}, execFn) {
        super(parent, options);
        this.name = name;
        this.args = args;

        this.execute = execFn;
    }

    getArgs () {
        return this.args;
    }

    getArg (name) {
        return this.args[name];
    }
}

module.exports = ApplicationNode;