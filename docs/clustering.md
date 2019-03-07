# Clustering

There are some different concerns here.  First we describe the clustering of the underlying storage, Cassandra.
Next is the clustering of metrictank itself, which is really two separate features (partitioning for horizontal scalability, and replication for high availability),
described separately below, though they can be combined.

## Underlying storage

For clustering (in all senses of the word) of the underlying storage, we of course can simply rely on Cassandra. It has built-in replication and data partitioning to assure both HA and load balancing.

## Metrictank horizontal scaling (partitioning)

If single nodes are incapable of handling your volume of metrics, you will want to split up your data in shards or partitions, and have multiple metrictank instances each handle a portion of the stream (e.g. one or a select few shards per instance).

* When ingesting data via kafka, you can simply use Kafka partitions.  The partition config setting for the plugin will control which instances consume which partitions.
* When using carbon, you can route data by setting up the carbon connections manually (possibly using a relay). It is important that you set the configuration of the plugin to reflect how you actually route your traffic.

Any instance can serve reads for data residing anywhere in (and spread throughout) the cluster.
Instances join the cluster by announcing themselves to the `peers` set in their configuration.
Any pre-existing instances part of the cluster will receive the announcement directly, or via gossip.
An instance will regularly poll the health of other nodes and involve healthy peers if they host data we might not have locally.

Please see "Metrictank horizontal scaling plus high availability" below for a caveat.


## Metrictank for high availability (replication)

Metrictank achieves redundancy and fault tolerance by running multiple instances which receive identical inputs.
One of the instances needs to have the primary role, which means it saves data chunks to Cassandra.   The other instances are secondaries.

Configuration of primary vs secondary:

* statically in the [cluster section of the config](https://github.com/grafana/metrictank/blob/master/docs/config.md#clustering) for each instance.
* dynamically (see [http api docs](https://github.com/grafana/metrictank/blob/master/docs/http-api.md)) should your primary crash or you want to shut it down.

### Clustering transport and synchronisation

The primary sends out persistence messages when it saves chunks to Cassandra.  These messages simply detail which chunks have been saved.
If you want to be able to promote secondaries to primaries, it's important they have been ingesting and processing these messages, so that the moment they become primary,
they don't start saving all the chunks it has in memory, which could be a significant sudden load on Cassandra.

Metrictank supports 1 transports for clustering: kafka, configured in the [clustering transports section in the config](https://github.com/grafana/metrictank/blob/master/docs/config.md#clustering-transports)

Instances should not become primary when they have incomplete chunks (though in worst case scenario, you might
have to do just that).  So they expose metrics that describe when they are ready to be upgraded.
Notes:
* all instances receive the same data and have a full copy in this case.
* primary role controls who writes *data* chunks to cassandra. *metadata* is updated by all instances independently and redundantly, see [metadata](https://github.com/grafana/metrictank/blob/master/docs/metadata.md).

### Promoting a secondary to primary

If the primary crashed, or you want to take it down for maintenance you will need to upgrade a secondary instance (the "candidate") to primary status.
This procedure needs to be done carefully:

1) assure there is no primary running.  Demote the primary to secondary if it's running. Make sure all the persistence messages made it through the clustering transport into the 
candidate. Synchronisation over the transport is usually near-instant so waiting a few seconds is usually fine.  But if you want to be careful, assure that the primary stopped sending persistence messages (via the dashboard), and verify that the candidate caught up with the transport (by monitoring the consumption delay from the transport).

2) pick a node that has `cluster.promotion_wait` at zero.  This means the instance has been consuming data long enough to have full data for all the recent chunks that will have to be saved - assuming the primary was able to persist its older chunks to Cassandra.  If that's not the case, just pick the instance that has been consuming data the longest or has a full working set in RAM (e.g. has been for longer than [`chunkspan * numchunks`](https://github.com/grafana/metrictank/blob/master/docs/config.md#data).  (note: when a metrictank process starts, it first does a few maintenance tasks before data consumption starts, such as filling its index when needed)
The `cluster.promotion_wait` is automatically determined based on the largest chunkSpan (typically your coarsest aggregation) and the current time.  From which an instance can derive when it's ready.

3) open the Grafana dashboard and verify that the secondary is able to save chunks 

## Combining metrictank's horizontal scaling plus high availability.

If you use both the partitioning (for write load sharding) and replication (for fault tolerance) it is important that the replicas consume the same partitions, and hence, contain the same data.
We call the "set of consumed shards" a shard group.

For example, you could set up your nodes like so:

node       | A   | B   | C   | D   |
| -------- | --- | --- | --- | --- |
partitions | 0,1 | 0,1 | 2,3 | 2,3 |


Let's say A and C are primaries.  When A fails, B can take over as primary. But note that it will have to process 200% of the read load for data in these partitions.

The counter example is something like:

node       | A   | B   | C   | D   |
| -------- | --- | --- | --- | --- |
partitions | 0,1 | 0,2 | 1,3 | 2,3 |

This would offer better load balancing should node A fail (B and C will each take over a portion of the load), but will require making primary status a per-partition concept.
Hence, this is currently **not supported**.

## Priority and ready state

Priority is a measure of how in-sync a metrictank process is, expressed in seconds.

| input plugin  | priority                 |
| ------------- | ------------------------ |
| carbon-in     | 0                        |
| prometheus-in | 0                        |
| kafka-mdm-in  | estimate of consumer lag |

When the input plugin is not sure, or not started yet priority is 10k (2.8 hours)

* the priority value is gossiped to all peers in a sharding cluster
* To satisfy queries from users, requests are fanned out across the cluster across all ready instances (see below), favoring lower priority instances.
* The priority can be inspected via http endpoints like `/cluster`, `/priority`, `/node` or via the dashboard/metrics.

Readyness or "ready state":

(whenever we say "ready", "ready state" we mean the value taking into account priority, not the internal NodeState, as explained below)

* indicates whether an instance is considered ready to satisfy data requests
* refuses data or index requests when not ready
* Can be checked via the `/` http endpoint. [more info](http-api.md#get-app-status)
* Can control the GC setting via the `cluster.gc-percent-not-ready` setting.

A node is ready when all of the following are true:
* priority does not exceed the `cluster.max-priority` setting, which defaults to 10.
* its internal NodeState is ready, which happens:
  * for primary nodes, immediately after startup (loading index, starting input plugins, etc)
  * for secondary nodes, "warm-up-period" after startup.

Special cases:
* the `/node` and `/cluster` endpoint shows the internal state of the node, including the internal NodeState.
* what is gossiped across the cluster is also the full internal node state (including NodeState, priority, etc)
* The `cluster.self.state.ready.gauge1` metric is also the internal NodeState, whereas the `cluster.total.state` metrics use the normal ready state.

## Caveats

If you get the following error:
```
2016/11/23 09:27:34 [metrictank.go:334 main()] [E] failed to initialize cassandra. java.lang.RuntimeException: java.util.concurrent.ExecutionException: org.apache.cassandra.exceptions.ConfigurationException: Column family ID mismatch (found 103be8f0-b15f-11e6-92df-7964cb8cd63c; expected 103902c0-b15f-11e6-92df-7964cb8cd63c)
```

This is because you start multiple metrictanks concurrently and they all try to initialize cassandra.
Enable the `create-keyspace` on only one node, or leave it enabled for all, but only start the others after successfully starting the first.

## Other
- use min-available-shards to control how many unavailable shards are tolerable
- note: currently if a shard fails, it doesn't retry other instance in the same request
