
# WIP

Cortex is a fault-tolerant(raft) alerts correlation engine. 

Alerts are accepted as a standard cloudevents.io event. Site 24x7 and Icinga integration converters are also provided.

It aggregates similar events over a time window using a regex matcher and executes a JS(ES6) script. The script contains the correlation logic which can further create incidents or alerts. The JS environment is limited and is achieved by embedding k6.io javascript interpreter(https://docs.k6.io/docs/modules). This is an excellent module built on top of https://github.com/dop251/goja


The aggregation of events in a bucket is done by writing a rule: 

```
{
    Title: ID: test-rule-id-1,
    EventTypes:[myntra.prod.icinga.check_disk,myntra.prod.site247.cart_down],
    ScriptID: myscript.js,
    Dwell:4000,
    DwellDeadline:3800,
    MaxDwell:8000,
    HookEndpoint: http://localhost:3000/testrule,
    HookRetry: 2,
}
```

where 

*EventTypes* is the pattern of events to put in a bucket(collection of cloudevents) associated with the rule.

*Dwell* is the wait duration since the first matched event.

### Event Flow:

Steps: 

1. **Match** : alert --> (convert from site 24x7/icinga ) --> (match rule) --> **Collect**
2. **Collect** --> (add to the rule bucket which *dwells* around until the configured time) -->  **Execute**
3. **Execute** --> (flush after Dwell period) --> (execute configured script) --> *Post*
4. **Post** --> (if result is set from script, post the result to the HookEndPoint or post the bucket itself if result is nil)


