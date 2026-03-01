# Kubernetes Standards

## Controllers

Generally, you should always try to use the cache wherever possible and make your controller able to tolerate stale cache reads by the following means:

### Guard status updates with a state check

Before writing a status or phase to a resource, compare the desired state with the current state. Skip the write if nothing has changed.

### Leverage optimistic locking: use deterministic naming for objects you create (this is what the Deployment controller does).

- Leverage optimistic locking / concurrency control of the API server: send updates/patches with the last-known resourceVersion from the cache (see below). This will make the request fail, if there were concurrent updates to the object (conflict error). This indicates that the controller has operated on stale data and might have made wrong decisions. In this case, let the controller handle the error with exponential backoff (simply return an error in your reconciler). This will make the controller eventually consistent.
- Track the actions your controller takes, e.g., when creating objects with generateName (this is what the ReplicaSet controller does [3]). The actions can be tracked in memory and repeated if the expected watch events don't occur after a given amount of time.
- Always try to write controllers with the assumption that data will only be eventually correct and can be slightly out-of-date (even if read directly from the API server!).
- If there is already some other code that needs a cache (e.g., a controller watch), reuse it instead of doing extra direct reads.
- Don’t read an object again if you just sent a write request. Write requests (Create, Update, Patch and Delete) don't interact with the cache. Hence, use the current state that the API server returned (filled into the passed in-memory object), which is basically a "free direct read", instead of reading the object again from a cache, because this will probably set back the object to an older resourceVersion.
- If it’s not possible to follow these rules in one of your controllers — which might be fine, but should rather be an exception — consider the following points:

If you are concerned about the impact of the resulting cache, try to minimize that by using filtered or metadata-only watches. If watching and caching an object type is not feasible, for example because there will be a lot of updates, and you are only interested in the object every ~5m, or because it will blow up the controllers memory footprint, fallback to a direct read. This can either be done by disabling caching the object type generally or doing a single request via an APIReader. In any case, bear in mind that every direct API call results in a quorum read from etcd, which can be costly in a heavily-utilized cluster and impose significant scalability limits. Thus, always try to minimize the impact of direct calls by filtering results by namespace or labels, limiting the number of results and/or using metadata-only calls.
