# HTTP api

This documents endpoints aimed to be used by users. For internal clustering http endpoints, which may change, refer to the source.

- Note that some of the endpoints rely on being a fed a proper Org-Id.  See [Multi-tenancy](https://github.com/grafana/metrictank/blob/master/docs/multi-tenancy.md).

- For GET requests, any parameters not specified as a header can be passed as an HTTP query string parameter.

## Get app status

```
GET /
POST /
```

returns:

* `200 OK` if the node is [ready](clustering.md#priority-and-ready-state)
* `503 Service not ready` otherwise.

#### Example


```bash
curl "http://localhost:6060"
```


## Walk the metrics tree and return every metric found that is visible to the org as a sorted JSON array

```
GET /metrics/index.json
POST /metrics/index.json
```

* header `X-Org-Id` required

Returns metrics stored under the given org, as well as public data (see [multi-tenancy](https://github.com/grafana/metrictank/blob/master/docs/multi-tenancy.md))

#### Example


```bash
curl -H "X-Org-Id: 12345" "http://localhost:6060/metrics/index.json"
```

## Find all metrics visible to the given org

```
GET /metrics/find
POST /metrics/find
```

* header `X-Org-Id` required
* query (required): can be an id, and use all graphite glob patterns (`*`, `{}`, `[]`, `?`)
* format: json, treejson, completer, pickle, or msgpack. (defaults to json)
* jsonp

Returns metrics which match the query and are stored under the given org or are public data (see [multi-tenancy](https://github.com/grafana/metrictank/blob/master/docs/multi-tenancy.md))
the completer format is for completion UI's such as graphite-web.
json and treejson are the same.

#### Example

```bash
curl -H "X-Org-Id: 12345" "http://localhost:6060/metrics/find?query=statsd.fakesite.counters.session_start.*.count"
```

## Deleting metrics

This will delete any metrics (technically metricdefinitions) matching the query from the index.
Note that the data stays in the datastore until it expires.
Should the metrics enter the system again with the same metadata, the data will show up again.

```
POST /metrics/delete
```

* header `X-Org-Id` required
* query (required): can be a metric key, and use all graphite glob patterns (`*`, `{}`, `[]`, `?`)

#### Example

```bash
curl -H "X-Org-Id: 12345" --data query=statsd.fakesite.counters.session_start.*.count "http://localhost:6060/metrics/delete"
```

## Graphite query api

This is the early beginning of a graphite-web replacement. It can return JSON, pickle or messagepack output
This section of the api is **very early stages**.  Your best bet is to use graphite in front of metrictank, for now.

```
GET /render
POST /render
```

* header `X-Org-Id` required
* maxDataPoints: int (default: 800)
* target: mandatory. one or more metric names or patterns, like graphite.
  note: **no graphite functions are currently supported** except that
  you can use `consolidateBy(id, '<fn>')` or `consolidateBy(id, "<fn>")` where fn is one of `avg`, `average`, `min`, `max`, `sum`. see
  [Consolidation](https://github.com/grafana/metrictank/blob/master/docs/consolidation.md)
* from: see [timespec format](#tspec) (default: 24h ago) (exclusive)
* to/until : see [timespec format](#tspec)(default: now) (inclusive)
* format: json, msgp, pickle, or msgpack (default: json)
* process: all, stable, none (default: stable). Controls metrictank's eagerness of fulfilling the request with its built-in processing functions
  (as opposed to proxing to the fallback graphite).
  - all: process request without fallback if we have all the needed functions, even if they are marked unstable (under development)
  - stable: process request without fallback if we have all the needed functions and they are marked as stable.
  - none: always defer to graphite for processing.

  If metrictank doesn't have a requested function, it always proxies to graphite, irrespective of this setting.

Data queried for must be stored under the given org or be public data (see [multi-tenancy](https://github.com/grafana/metrictank/blob/master/docs/multi-tenancy.md))

#### Example

```bash
curl -H "X-Org-Id: 12345" "http://localhost:6060/render?target=statsd.fakesite.counters.session_start.*.count&from=3h&to=2h"
```

## Get Cluster Status

```
GET /node
```

returns a json document with the following fields:

* "name": the node name
* "primary": whether the node is a primary node or not
* "primaryChange": timestamp of when the primary state last changed
* "version": metrictank version
* "state": whether the node is [ready](clustering.md#priority-and-ready-state) to handle requests or not
* "stateChange": timestamp of when the state last changed
* "started": timestamp of when the node started up

#### Example

```bash
curl "http://localhost:6060/node"
```

## Set Cluster Status

```
POST /node
```

parameter values :

* `primary true|false`

Sets the primary status to this node to true or false.

#### Example

```bash
curl --data primary=true "http://localhost:6060/node"
```

## Analyze instance priority

```
GET /priority
```

Gets all enabled plugins to declare how they compute their
priority scores.

#### Example

```bash
curl -s http://localhost:6060/priority | jsonpp
[
    "carbon-in: priority=0 (always in sync)",
    {
        "Title": "kafka-mdm:",
        "Explanation": {
            "Explanation": {
                "Status": {
                    "0": {
                        "Lag": 1,
                        "Rate": 26494,
                        "Priority": 0
                    },
                    "1": {
                        "Lag": 2,
                        "Rate": 24989,
                        "Priority": 0
                    }
                },
                "Priority": 0,
                "Updated": "2018-06-06T11:56:24.399840121Z"
            },
            "Now": "2018-06-06T11:56:25.016645631Z"
        }
    }
]

## Cache delete

```
GET /ccache/delete
POST /ccache/delete
```

* `X-Org-Id`: required
* patterns: one or more query (glob) patterns. Use `**` to mean "all data" (full reset)
* expr: tag expressions
* propagate: whether to propagate to other cluster nodes. true/false

Remove chunks from the cache for matching series, or wipe the entire cache

#### Example

```bash
curl -v -X POST -d '{"propagate": true, "orgId": 1, "patterns": ["**"]}' -H 'Content-Type: application/json' http://localhost:6060/ccache/delete
```

## Misc

### Tspec

The time specification is used throughout the http api and it can be any of these forms:

* now: current time on server
* any integer: unix timestamp
* `-offset` or `offset` gets interpreted as current time minus offset.
  Where `offset` is a sequence of one or more `<num><unit>` where unit is one of:

	- ``, `s`, `sec`, `secs`, `second`, `seconds`
	- `m`, `min`, `mins`, `minute`, `minutes`
	- `h`, `hour`, `hours`
	- `d`, `day`, `days`
	- `w`, `week`, `weeks`
	- `mon`, `month`, `months`
	- `y`, `year`, `years`

* datetime in any of the following formats: `15:04 20060102`, `20060102`, `01/02/06`

