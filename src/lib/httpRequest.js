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

var DEF_EXPORT_VAR = function () {
    return {
        "effective_caller_id_name": "callerIdName",
        "owner_caller_id_number": "callerIdOwner"
    };
};

var DEF_HEADERS = function () {
    return {
        "Content-Type":"application/json"
    }
};

var DEF_DATA = function() {
    return {
        "callerIdNumber": "${Caller-Caller-ID-Number}"
    };
};

client.on('error', function (err) {
    log.error(err);
});


module.exports = function (parameters, router, cb) {

    if (!parameters['url']) {
        cb(new Error('Bad request'));
        return;
    };
    
    var _parseRequest = function (dataRequest) {
        try {
            var jsonData = JSON.parse(dataRequest);
            log.debug(jsonData);
            for (var key in exportVariables) {
                if (!exportVariables.hasOwnProperty(key)) continue;

                if (jsonData.hasOwnProperty(exportVariables[key]) && jsonData[exportVariables[key]]) {
                    router._set({
                        "setVar": "all:" + key + "=" + jsonData[exportVariables[key]]
                    });
                };
            };
        } catch (e) {
            log.error(e.message);
        } finally {
            cb();
        };
    };

    var method = parameters['method'] || 'post',
        exportVariables = parameters['exportVariables'] || DEF_EXPORT_VAR(),
        headers = parameters['headers'] || DEF_HEADERS();


    var webArgs = {
        data: parameters['data'] || DEF_DATA(),
        headers: headers,
        requestConfig:{
            timeout: 1000, //request timeout in milliseconds
            keepAlive: false //Enable/disable keep-alive functionalityidle socket.
        },
        responseConfig: {
            timeout: 1000 //response timeout
        }
    };
    method = method.toLowerCase();

    for (var key in webArgs.data) {
        if (!webArgs.data.hasOwnProperty(key)) continue;

        if (/^\$\$\{\W*\w*/.test(webArgs.data[key])) {
            webArgs.data[key] = router.getGlbVar(webArgs.data[key].replace(/\$|\{|}/g, ''));
        } else if (/^\$\{\W*\w*/.test(webArgs.data[key])) {
            webArgs.data[key] = router.getChnVar(webArgs.data[key].replace(/\$|\{|}/g, ''));
        };
    };

    var req;
    if (method == METHODS.GET) {
        delete webArgs.data;
        req = client.get(parameters['url'], webArgs, _parseRequest);
    } else if (method == METHODS.POST) {
        req = client.post(parameters['url'], webArgs, _parseRequest);
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