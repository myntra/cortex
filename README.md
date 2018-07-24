
# experimental

cortex is a HA cloudevents.io aggregator


```
{
    "cloudEventsVersion" : "0.1",
    "eventType" : "servicename.subsystem.metric",
    "source" : "icinga", 
    "eventID" : "C234-1234-1234",
    "eventTime" : "2018-04-05T17:31:00Z",
    "extensions" : {
      "comExampleExtension" : "value"
    },
    "contentType" : "application/json",
    "data" : {
        "key" : "val",
    }
}

```

## EventType

Event Types are of the format:

```
<service-name>.<monitoring-system-name>.<metric-name>
```

## Rule

A rule is an array of related eventypes.

```{
    rule: ["service1.site24x7.",service2.icinga.metric*.businessname.productname.x, service2.icinga.*.businessname.productname.x]
    endpoint: "http://localhost:8080/correlate/search/up,
    waitWindow: 30000,
    slideWindow: 1000,
    snoozeWindow: 2000,
}


{
    rule: ["service1.site24x7.deadui",service1.icinga.*,service2.icing.metric]
    endpoint: "http://localhost:8080/correlate/service2,
}
```

## RuleBucket

A rulebucket is a collection of related cloudevents.

```
[]event{}
```

This bucket is posted to the remote endpoint on timeout and removed from the aggregator.


## AlertHandlers

An AlertHandler accepts alerts in various formats and stores them as cloudevents.

e.g: site 24x7, icinga etc.

path: /site24x7/alert


Notes:
1. Sliding Time Window.
2. CorrelationService design: EventRegistry.
3. RuleCreator 