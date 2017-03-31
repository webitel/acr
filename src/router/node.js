/**
 * Created by igor on 29.03.17.
 */

"use strict";

const BaseNode = require('./baseNode');

class Node extends BaseNode {
    constructor (parent) {
        super(parent);
        this.children = [];
        this.position = 0;
    }

    get length () {
        return this.children.length;
    }

    add (node) {
        this.children.push(node)
    }

    first () {
        this.position = 0;
    }

    isLast () {
        return this.children.length === (this.position + 1);
    }

    next () {
        return this.children[this.position++];
    }
}

module.exports = Node;