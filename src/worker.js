var esl = require('modesl'),
    log = require('./lib/log')(module);

process.work_port = process.env['WORKER_PORT'];
log.info('start port: ' + process.env['WORKER_PORT']);

var esl_server = new esl.Server({host: '10.10.10.25', port: process.env['WORKER_PORT'], myevents:false}, function(){
    log.info("ESL server is up port " + this.port);
});

esl_server.on('connection::ready', function(conn, id) {
    conn.on('error', function (error) {
        log.error(error);
    });

    log.trace('new call ' + id);
    conn.call_start = new Date().getTime();
    //log.log(conn.channelData.serialize('plain'));
    conn.execute('answer');
    conn.execute('echo', function(){
        log.trace('echoing');
    });

    conn.on('esl::end', function(evt, body) {
        this.call_end = new Date().getTime();
        var delta = (this.call_end - this.call_start) / 1000;
        log.trace("Call duration " + delta + " seconds");
    });
});

esl_server.on('error', function (err) {
    log.error(err);
});