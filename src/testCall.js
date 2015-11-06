/**
 * Created by i.navrotskyj on 27.01.2015.
 * ./sipp 10.10.10.200:5070 -i 10.10.10.133 -s 333 -d 5s -l 50 -aa -mi 10.10.10.133 -rtp_echo -nd -r 10
 */
var Esl = require('modesl'),
    log = require('./lib/log');

var calle = '101@10.10.10.144',
    ext = '111';
var i = 0;

var esl = new Esl.Connection('10.10.10.160', 8021, 'ClueCon', function() {
//var esl = new Esl.Connection('10.10.10.145', 8021, 'ClueCon', function() {
        console.info('Connect freeSWITCH - OK');
    setInterval(function () {
        //if (i==20) return;
        esl.bgapi("originate sofia/external/111@it-sfera.com.ua:5080 &echo()", function (res) {
            //console.log("Call: %d", i++);
            //if (i == 100) process.exit(0);
        });

    }, 40);
    //return

    //for (var i = 0; i < 50; i++) {
    //        esl.bgapi(('originate sofia/external/111@194.44.216.235:5080 ' + ext +
    //        ' xml default ' + calle + ' ' + calle), function (res) {
    //            console.log("Call: %d", i);
    //        });
    //    }
    esl.subscribe(['CHANNEL_DESTROY', "CHANNEL_CREATE"]);
});
/// hupall

esl.on('esl::event::CHANNEL_CREATE::*', function (event) {
    if (event.getHeader('Channel-Name') == 'sofia/external/111@it-sfera.com.ua:5080') {
        console.log('NEW CALL: ' + (i++))
    }
}) ;

esl.on('esl::event::CHANNEL_DESTROY::*', function (event) {
    //if (event.getHeader('variable_node_call') == 'test' ) {
        console.log('END CALL: ' + (i--));
    //}
}) ;