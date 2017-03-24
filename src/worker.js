const path = require('path');

global.__appRoot = path.resolve(__dirname);


const acr = require('./acr'),
    log = require('./lib/log')(module)
    ;


process.on('uncaughtException', function (err) {
    log.error('UncaughtException:', err.message);
    log.error(err.stack);
    process.exit(1);
});

if (typeof gc == 'function') {
    setInterval( () => {
        console.log('----------------------- GC -----------------------');
        gc();
    }, 5000)
}