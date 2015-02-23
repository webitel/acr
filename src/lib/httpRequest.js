/**
 * Created by i.navrotskyj on 17.02.2015.
 */
var Client = require('node-rest-client').Client,
    client = new Client(),
    //EventEmitter2 = require('eventemitter2').EventEmitter2,
    log = require('./log')(module);

var METHODS = {
    GET: "get",
    POST: "post",
    PUT: "put",
    DELETE: "delete"
};

var DEF_EXPORT_VAR = {
    "effective_caller_id_name": "callerIdName",
    "owner_caller_id_number": "callerIdOwner"
};

var DEF_HEADERS = {
    "Content-Type":"application/json"
};

var DEF_DATA = {
    "callerIdNumber": "${caller_id_number}"
};

client.on('error', function (err) {
    log.error(err);
});

module.exports = function (parameters, router, cb) {

    function parseRequest (data, response) {
        try {
            var jsonData = JSON.parse(data);
            log.debug(jsonData);
            for (var key in exportVariables) {
                if (jsonData.hasOwnProperty(exportVariables[key]) && jsonData[exportVariables[key]]) {
                    router._set({
                        "set": "all:" + key + "=" + jsonData[exportVariables[key]]
                    });
                };
            };
        } catch (e) {
            log.error(e.message);
        } finally {
            cb();
        };
    };

    if (!parameters['url']) {
        cb(new Error('Bad request'));
        return;
    };

    var method = parameters['method'] || 'post',
        method = method.toLowerCase(),
        exportVariables = parameters['exportVariables'] || DEF_EXPORT_VAR,
        headers = parameters['headers'] || DEF_HEADERS,
        data = parameters['data'] || DEF_DATA;

    if (typeof data == "object") {
        for (var key in data) {
            if (/^\$\$\{\W*\w*/.test(data[key])) {
                data[key] = router.getGlbVar(data[key].replace(/\$|\{|}/g, ''));
            } else if (/^\$\{\W*\w*/.test(data[key])) {
                data[key] = router.getChnVar(data[key].replace(/\$|\{|}/g, ''));
            };
        };
    };

    var webArgs = {
        data: data,
        headers: headers,
        requestConfig:{
            timeout: 1000, //request timeout in milliseconds
            keepAlive: false //Enable/disable keep-alive functionalityidle socket.
        },
        responseConfig: {
            timeout: 1000 //response timeout
        }
    };
    console.dir(webArgs);
    var req;
    if (method == METHODS.GET) {
        delete webArgs.data;
        req = client.get(parameters['url'], webArgs, parseRequest);
    } else if (method == METHODS.POST) {
        req = client.post(parameters['url'], webArgs, parseRequest);
    } else {
        log.error('Bad parameters method');
        cb();
        return;
    };

    req.on('requestTimeout',function(req){
        log.warn("request has expired");
        req.abort();
        cb();
    });

    req.on('error',function(err){
        log.error(err.message);
        cb();
    });

    req.on('responseTimeout',function(){
        log.warn("response has expired");
    });
};