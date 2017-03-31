/**
 * Created by igor on 29.03.17.
 */

"use strict";

class BaseNode {
    constructor (parent, options = {}) {
        this._depth = 0;
        this._parent = parent;
        this.idx = 0;
        if (parent) {
            this.idx = parent.length;
            this._depth = parent.depth + 1;
        }

        [
            this.break = false,
            this.async = false,
            this.dump = false
        ] = [
            options.break,
            options.async,
            options.dump
        ]
    }

    get depth () {
        return this._depth;
    }

    getParent () {
        return this._parent;
    }
}

module.exports = BaseNode;