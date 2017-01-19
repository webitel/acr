/**
 * Created by igor on 19.01.17.
 */

"use strict";

const os = require('os')
    ;


const Service = module.exports = {

    allStats: () => {
        return {
            "version": process.env['VERSION'] || '',
            "nodeMemory": Service.memoryUsage(),
            "processId": process.pid,
            "processUpTimeSec": process.uptime(),
            "system": Service.osInfo(),
            "crashCount": process.env['CRASH_WORKER_COUNT'] || 0,
            "nodeVersion": process.version
        }
    },

    osInfo: () => {
        return {
            "totalMemory": os.totalmem(),
            "freeMemory": Service.freeMemory(),
            "platform": os.platform(),
            "name": os.type(),
            "architecture": os.arch()
        };
    },

    freeMemory: () => {
        return os.freemem();
    },

    memoryUsage: () => {
        let memory = process.memoryUsage();
        return {
            "rss": memory.rss,
            "heapTotal": memory.heapTotal,
            "heapUsed": memory.heapUsed
        }
    },

    cpuInfo: () => {
        let res = {},
            cpus = os.cpus()
            ;
        for(var i = 0, len = cpus.length; i < len; i++) {
            res['CPU' + i] = {};
            var cpu = cpus[i], total = 0;
            for(var type in cpu.times)
                total += cpu.times[type];

            for(type in cpu.times)
                res['CPU' + i][type] = Math.round(100 * cpu.times[type] / total)
        }
        return res;
    }
};