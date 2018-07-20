var revenueEvent = require("./revenueEvent")

function createCloudEvent(data) {
    var RCACloudEvent = {
        "cloudEventsVersion": "0.1",
        "eventType": "RCAObject",
        "source": "EventCorrelator",
        "eventID": "",
        "eventTime": "",
        "contentType": "application/json",
        "data": {
        }
    }
    
    RCACloudEvent.eventID = generateRandomID()
    RCACloudEvent.eventTime = currentTimeFormatter()
    RCACloudEvent.data = data

    return RCACloudEvent
}


function generateRandomID() {
    return Math.random().toString(36).substring(7);
}

function currentTimeFormatter() {
    var d = new Date,
        dformat = ("00" + (d.getMonth() + 1)).slice(-2) + "/" +
        ("00" + d.getDate()).slice(-2) + "/" +
        d.getFullYear() + " " +
        ("00" + d.getHours()).slice(-2) + ":" +
        ("00" + d.getMinutes()).slice(-2) + ":" +
        ("00" + d.getSeconds()).slice(-2)

    return dformat;
}


var ruleRoutes = {
    revenueEvent: function (res) {
        return createCloudEvent(revenueEvent.revenueEvent(res.body))
    },
    default: function (res) {
        console.log("RuleType could not be matched")
        return createCloudEvent({"error":"Rule type couldnot by found"})
    }


};

exports.ruleRoutes = ruleRoutes;