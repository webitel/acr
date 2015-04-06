var util = require('util');

var ob = {
    "name": 1,
    "2012": {
        "2": 3
    }
};

console.log(util.format('%.name', ob));