/**
 * Created by igor on 27.03.17.
 */

"use strict";

const log = require(__appRoot + '/lib/log')(module),
    acr = require(__appRoot + '/acr'),
    Node = require('./node'),
    ApplicationNode = require('./applicationNode'),
    BreakNode = require('./breakNode'),
    SwitchNode = require('./switchNode'),
    ConditionNode = require('./conditionNode');


const MAX_GOTO_COUNTER = 100;

class Iterator {
    constructor (callflow, acr) {
        this.tags = new Map();
        this.functions = new Map();
        this._current = new Node(null);

        this.getExecuteFunction = (name) => acr.getApplication(name);
        this.getAcr = () => acr;

        this._gotoCounter = 0;

        if (callflow instanceof Array) {
            this._rootCount = callflow.length - 1;
            this._parseCallFlow(callflow, this._current);
        }

        //this.bigData = new Array(1e6).join('XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX\n');


        // const test = (n) => {
        //     if (!n) {
        //         console.log("END!!!!!!!!!!!!!!!!!");
        //         return;
        //     }
        //
        //     n.execute(this);
        //
        //     if (n.name === 'goto') {
        //         this.goto(n.getArgs())
        //     }
        //
        //     test(this.next() || this.getParent());
        //
        // };
        //
        // test(this.next() || this.getParent());

    }

    goto (tagName) {
        // TODO add support old version from root number ?
        const foundTagMap = this.tags.has(tagName);
        let gotoApp;
        if (!foundTagMap && typeof tagName !== 'number') {
            return false;
        } else if (!foundTagMap && typeof tagName === 'number') {
            // TODO remove next version
            log.debug(`Deprecated goto from root position ${tagName}`);
            if (this._rootCount < tagName && tagName > this._rootCount) {
                log.warn(`Deprecated goto not found index ${tagName}`);
                return false;
            }

            while (this.getParent()) {

            }
            gotoApp = this._current.children[tagName];

            if (!gotoApp) {
                log.warn(`Deprecated goto not found index ${tagName} by skip bad applications... move to end!!!`);
                gotoApp = this._current.children[this._current.children.length - 1];

            }
        } else {
            gotoApp = this.tags.get(tagName);
        }

        
        this._current.first();
        this.setRoot(gotoApp.getParent());
        this._current.position = gotoApp.idx;
        
        if (this._current._parent) {
            this._current._parent.position = this._current.idx + 1;
        }
        this._gotoCounter++;
        return true;
    }

    getFunction (name) {
        return this.functions.get(name);
    }

    setRoot (root) {
        if (root instanceof Node) {
            this._current = root;
        } else {
            throw root;
        }
    }

    getParent () {
        const parent = this._current.getParent();
        this._current.first();
        if (!parent)
            return null;

        this._current = parent;
        return this.next() || this.getParent();
    }

    _checkTag (app, node) {
        if (app.tag != null) {
            this.tags.set(app.tag, node);
        }
    }


    _parseCallFlow (obj, root) {
        if (obj instanceof Array) {
            for (let app of obj) {
                const {name, args} = getApplicationParameters(app);

                if (!name) {
                    log.warn(`Skip bad application: `, app);
                    continue;
                }

                let node = null;

                switch (name) {

                    case 'if':
                        node = new ConditionNode(root, app.if, args);
                        root.add(node);

                        if (app.if.then instanceof Array) {
                            this._parseCallFlow(app.if.then, node.getThenNode())
                        }
                        if (app.if.else instanceof Array) {
                            this._parseCallFlow(app.if.else, node.getElseNode())
                        }
                        break;

                    case 'break':
                        node = new BreakNode(root);
                        root.add(node);
                        break;

                    case 'function':
                        if (!app.function.name || !(app.function.actions instanceof Array)) {
                            log.warn(`Bad function parameters: `, app);
                            continue;
                        }
                        node = new Iterator(app.function.actions, this.getAcr()); // new FunctionNode(app.function.name, app.function.actions);
                        this.functions.set(app.function.name, node);
                        continue;

                    case 'switch':
                        node = new SwitchNode(root, app.switch, args);
                        root.add(node);
                        node._values.forEach( valueName => {
                            this._parseCallFlow(node.getCaseWorkFlow(valueName), node.getValueNode(valueName));
                        });

                        break;

                    default:
                        const execFn = this.getExecuteFunction(name);
                        if (typeof execFn !== 'function') {
                            log.warn(`Skip bad application: `, app);
                            continue;
                        }
                        node = new ApplicationNode(root, name, app[name], args, execFn);
                        root.add(node);
                        this._parseCallFlow(app, root);
                }

                this._checkTag(app, node);
            }
        }
    }

    next () {
        if (this._gotoCounter >= MAX_GOTO_COUNTER) {
            log.error(`Cycle goto application: ${this._gotoCounter}`);
            return null;
        }

        return this._current.next();
    }
}


/**
 *
 * @param app
 * @returns {{name: null, args: {}}}
 */
const getApplicationParameters = (app) => {
    const result = {name: null, args: {}};

    if (!(app instanceof Object))
        return result;

    const propKeys = Object.keys(app);
    if (propKeys.length === 1) {
        result.name = propKeys[0];
    } else if (propKeys.length > 0) {
        for (let propName of propKeys){
            if (!result.name && propName !== 'break' && propName !== 'async' && propName !== 'tag' && propName !== 'dump') {
                result.name = propName;
            } else {
                result.args[propName] = app[propName]
            }
        }
    }

    if (app[result.name] == null)
        result.name = null;

    return result;
};

module.exports = Iterator;