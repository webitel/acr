// Include the cluster module
var cluster = require('cluster'),
    log = require('./lib/log')(module),
    conf = require('./conf'),
    port = parseInt(conf.get('server:ports')),
    ports = [port],
    _ports = {},
    _k = ports.length;

// Code to run if we're in the master process
if (cluster.isMaster) {
    if (process.env['ACR_COUNT']) {
        for (var i = 1, len = parseInt(process.env['ACR_COUNT']); i < len; i++) {
            ports.push(port + i);
        };
        _k = ports.length;
    };

    var debug = process.execArgv.indexOf('--debug') !== -1;
    cluster.setupMaster({
        execArgv: process.execArgv.filter(function(s) { return s !== '--debug' })
    });

    for (var i = 0; i < ports.length; i ++) {
        var new_worker_env = {};
        new_worker_env["WORKER_PORT"] = ports[i];
        _ports["p" + (i + 1)] = new_worker_env["WORKER_PORT"];

        if (debug) cluster.settings.execArgv.push('--debug=' + (5859 + i));
        cluster.fork(new_worker_env);
        if (debug) cluster.settings.execArgv.pop();
    }

    // Listen for dying workers
    cluster.on('exit', function (worker) {
        _k++;
        var new_worker_env = {};
        new_worker_env["WORKER_PORT"] = _ports["p" + worker.id];
        delete _ports["p" + worker.id];
        _ports["p" + _k] = new_worker_env["WORKER_PORT"];
        console.dir(_ports);

        // Replace the dead worker, we're not sentimental
        log.error('Worker ' + worker.id + ' died.');
        if (debug) cluster.settings.execArgv.push('--debug=' + (5859 + i));
        cluster.fork(new_worker_env);
        if (debug) cluster.settings.execArgv.pop();
    });

// Code to run if we're in a worker process
} else {
    require('./worker');
    log.trace('Worker ' + cluster.worker.id + ' running!');
}