
# WIP

**Cortex** is a fault-tolerant(raft) alerts correlation engine. 

Alerts are accepted as a standard cloudevents.io event. Site 24x7 and Icinga integration converters are also provided.

It collects similar events in a bucket over a time window using a regex matcher and then executes a JS(ES6) script. The script contains the correlation logic which can further create incidents or alerts. The JS environment is limited and is achieved by embedding k6.io javascript interpreter(https://docs.k6.io/docs/modules). This is an excellent module built on top of https://github.com/dop251/goja


The collection of events in a bucket is done by writing a rule: 

```json
{
	"title": "a test rule",
	"id": "test-rule-id-1",
	"eventTypes": ["acme.prod.icinga.check_disk", "acme.prod.site247.*"],
	"scriptID": "myscript.js",
	"dwell": 4000,
	"dwellDeadline": 3800,
	"maxDwell": 8000,
	"hookEndpoint": "http://localhost:3000/testrule",
	"hookRetry": 2
}
```



where:

*EventTypes* is the pattern of events to put in a bucket(collection of cloudevents) associated with the rule.

*Dwell* is the wait duration since the first matched event.


For this rule incoming events with `eventType` matching one of `eventTypes` will be put in the same bucket.

### Event Flow:

Steps: 

1. **Match** : alert --> (convert from site 24x7/icinga ) --> (match rule) --> **Collect**
2. **Collect** --> (add to the rule bucket which *dwells* around until the configured time) -->  **Execute**
3. **Execute** --> (flush after Dwell period) --> (execute configured script) --> *Post*
4. **Post** --> (if result is set from script, post the result to the HookEndPoint or post the bucket itself if result is nil)


