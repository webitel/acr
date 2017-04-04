/**
 * Created by igor on 04.04.17.
 */

"use strict";


// Include the cluster module
let cluster = require('cluster'),
    countWorker = parseInt(process.env['COUNT']) || 4,
    crashCount = 1;

if (isFinite(countWorker)) {
    if (countWorker === 0)
        countWorker = require('os').cpus().length
} else {
    countWorker = 1;
}

// Code to run if we're in the master process
if (cluster.isMaster) {

    for (var i = 0; i < countWorker; i++) {
        cluster.fork();
    }

    // Listen for dying workers
    cluster.on('exit', function (worker) {

        // Replace the dead worker, we're not sentiment
        console.error('Worker ' + worker.id + ' died.');
        cluster.fork({
            "CRASH_WORKER_COUNT": (crashCount++)
        });
    });

// Code to run if we're in a worker process
} else {
    require('./worker');
    console.info('Worker ' + cluster.worker.id + ' running!');
}