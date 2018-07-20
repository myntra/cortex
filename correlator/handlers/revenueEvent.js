//TODO 1. Snooze
//2. DeDup


var revenueEventsList = ["site24x7.deadui", "devapi.5xx.metric", "icinga.cart.memory.metric"]
var revenueEvents = {
    "site24x7.deadui": {
        "status": "down"
    },
    "devapi.5xx.metric": {
        "error_count": 1000
    },
    "icinga.cart.memoryused.metric": {
        "memory": 90
    },
    "matchCount": 2
}

function revenueEvent(data) {

    var eventBucket = data.eventBucket
    var matchCount = 0
    var RCAObject = {}
    RCAObject['events'] = []

    for (var i = 0; i < eventBucket.length; i++) {
        if (typeof revenueEvents[eventBucket[i].eventType] != 'undefined') {
            if (eventBucket[i].eventType == "site24x7.deadui") {
                console.log("Inside if if ", i, eventBucket[i].eventType, eventBucket[i].data.status)

                if (eventBucket[i].data.status == "down") {
                    matchCount += 1
                    RCAObject['events'].push({
                        "event": eventBucket[i].eventType,
                        "reason": "site24x7: Myntra website is down",
                        "source": eventBucket[i].source
                    })
                }
            } else if (eventBucket[i].eventType == "devapi.5xx.metric") {
                if (eventBucket[i].data.error_count > 1000) {
                    matchCount += 1
                    RCAObject['events'].push({
                        "event": eventBucket[i].eventType,
                        "reason": "Devapi is giving 5xx errors more than threshold",
                        "source": eventBucket[i].source
                    })
                }
            } else if (eventBucket[i].eventType == "icinga.cart.memoryused.metric") {
                if (eventBucket[i].data.memory > 90) {
                    matchCount += 1
                    RCAObject['events'].push({
                        "event": eventBucket[i].eventType,
                        "reason": "Cart is having memory used more than the threshold",
                        "source": eventBucket[i].source
                    })
                }
            }
        } else {
            console.log("Ignoring the event", typeof revenueEvents[eventBucket[i].eventType] != 'undefined', revenueEvents[eventBucket[i].eventType])
        }
    }

    if (matchCount >= revenueEvents.matchCount) {
        console.log("Sending the RCAObject. Incident has occured")
        return {"events":RCAObject['events'],"incident":true,"error":""}
    } else {
        console.log("Ignoring the Events. Not sufficient data")
        return {"events":null,"incident":false,"error":""}
    }


}

exports.revenueEvent = revenueEvent;