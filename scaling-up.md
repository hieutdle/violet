# How the operator scales up a datacenter

## Quick Overview

## Detailed Analysis

### 1. Update size by 2

The scaling up process in the Cassandra operator is triggered when the `spec.size` field of a CassandraDatacenter is increased.

`kubectl patch cassandradatacenter/dc1 --patch '{"spec":{"size":3}}' --type merge`

It sends a PATCH HTTP request to the Kubernetes API server targeting the `dc1` instance of the `CassandraDatacenter` **custom resource**.

The operator watches the Kubernetes API for changes to `CassandraDatacenter` resources.

Inside `internal/controller/cassandradatacenter_controller.go`:

```go
c := ctrl.NewControllerManagedBy(mgr).
    Named("cassandradatacenter_controller").
    For(&api.CassandraDatacenter{}, builder.WithPredicates(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.AnnotationChangedPredicate{})))
```

The line

```go
For(&api.CassandraDatacenter{}, builder.WithPredicates(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.AnnotationChangedPredicate{})))
```

responsible for setting up the controller to watch for changes to `CassandraDatacenter` resources

Then it trigger the `Reconcile` method of the `CassandraDatacenter` controller.

### 2. Set `ScalingUp` to `True`

This step primarily consists of two key phases (still in `internal/controller/cassandradatacenter_controller.go`):

- Determining Desired Cluster State (`CalculateRackInformation()`)
- Applying Desired Cluster State (`ReconcileAllRacks()`)

**Phase 1: Calculate Desired State (`CalculateRackInformation()`)**

Before scaling, the operator calculates the desired configuration for each rack in your Cassandra datacenter:

- Nodes per rack: Calculates how many total nodes each rack should have (`NodeCount`)
- Seed nodes per rack: Determines how many seed nodes each rack requires (`SeedCount`)

In our cases, when scaling a `CassandraDatacenter` from `1` to `3` nodes:

1.  Initial Parsing

    ```go
     nodeCount := int(rc.Datacenter.Spec.Size)  // This is now 3 (was 1)
     racks := rc.Datacenter.GetRacks()          // Get rack configuration
     rackCount := len(racks)                    // Number of racks defined
    ```

    First checks if the requested node count is valid.

    ```go
    if nodeCount < rackCount && rc.Datacenter.GetDeletionTimestamp() == nil {
     return fmt.Errorf("the number of nodes cannot be smaller than the number of racks")
    }
    ```

2.  Seed Node Calculation

    ```go
     seedCount := 3
     if nodeCount < 3 {
         seedCount = nodeCount
     } else if rackCount > 3 {
         seedCount = rackCount
     }
    ```

    In our example, where we're scaling from `1` to `3` nodes:

    - When `size` was `1`: `seedCount` would have been `1`
    - When `size` is `3`: `seedCount` becomes `3` (the default for `3+` nodes)

    If the user has not specified a seed count, it defaults to `1`.

3.  Node Distribution Across Racks
    Next, the function calculates how the nodes should be distributed across racks:

    ```go
     rackSeedCounts := api.SplitRacks(seedCount, rackCount)  // Distribute seeds
     rackNodeCounts := api.SplitRacks(nodeCount, rackCount)  // Distribute nodes
    ```

    Inside the `SplitRacks`:

    ```go
     func SplitRacks(nodeCount, rackCount int) []int {
        nodesPerRack, extraNodes := nodeCount/rackCount, nodeCount%rackCount
        var topology []int

        for rackIdx := 0; rackIdx < rackCount; rackIdx++ {
            nodesForThisRack := nodesPerRack
            if rackIdx < extraNodes {
                nodesForThisRack++
            }
            topology = append(topology, nodesForThisRack)
        }

        return topology
     }
    ```

    The `nodesPerRack = nodeCount/rackCount` distribute the nodes evenly: each rack gets at least `nodesPerRack` nodes.
    In our case, It is `topology = 3/1 = [3]`.

    `SplitRacks(8, 3)` case, you can skip to the next section if you don't want to see the details.

    ```go
     nodeCount = 8
     rackCount = 3

     nodesPerRack = nodeCount / rackCount = 2
     extraNodes   = nodeCount % rackCount = 2
     topology     = []  // empty slice
    ```

    ```go
    for rackIdx := 0; rackIdx < rackCount; rackIdx++ {
    ```

    This means: we'll go through `rackIdx = 0`, `1`, `2`.

    ```go
    // First Iteration (rackIdx = 0):
     nodesForThisRack = nodesPerRack = 2
     if rackIdx < extraNodes {  // 0 < 2 --> true
         nodesForThisRack++     // nodesForThisRack = 3
     }
     topology = append(topology, 3)  // topology = [3]

    // Second Iteration (rackIdx = 1):
    nodesForThisRack = 2
    if 1 < 2 --> true
        nodesForThisRack++  // nodesForThisRack = 3
    topology = append(topology, 3)  // topology = [3, 3]

    // Third Iteration (rackIdx = 2):
        nodesForThisRack = 2
    if 2 < 2 --> false
        // nothing added
    topology = append(topology, 2)  // topology = [3, 3, 2]
    ```

4.  Creating Rack Information Objects
    The function then iterates over the configured racks, and creates a `RackInformation` object for each rack

    ```go
    desiredRackInformation = append(desiredRackInformation, nextRack)
    ```

5.  Setting Up for StatefulSet Changes
    The function creates placeholder StatefulSet references and stores the calculated data:

    ```go
    statefulSets := make([]*appsv1.StatefulSet, len(desiredRackInformation))
    // The operator will fill this later as it generates the actual StatefulSets for reconciliation.
    rc.desiredRackInformation = desiredRackInformation
    rc.statefulSets = statefulSets
    // This stores the computed target state inside the ReconciliationContext (rc), so that downstream parts of the reconciliation process — especially ReconcileAllRacks() — can use it.
    ```

**Phase 2: Applying Desired State (`ReconcileAllRacks()`)**

Once the desired configuration is calculated, the operator reconciles the current state of the cluster with this desired state. It proceeds through a sequence of reconciliation steps. Each step is designed to address a specific aspect of the datacenter configuration:

1. Check StatefulSet Controller:

```go
   if recResult := rc.CheckStatefulSetControllerCaughtUp(); recResult.Completed() {
       return recResult.Output()
   }
```

And then bunch of other checks:

- Refreshes the status of the CassandraDatacenter
- Verifies configuration secrets are properly set up
- Creates StatefulSets for racks if they don't exist
- Ensures proper labels on all rack resources
- Handles any ongoing node decommissioning operations
- Sets up authentication secrets
- Handles stopped datacenter state
- ...

But the key for scaling is:

`CheckRackScale()`:

```go
   if recResult := rc.CheckRackScale(); recResult.Completed() {
       return recResult.Output()
   }
```

And then after that still a bunch of other checks.

But we will focus on `CheckRackScale()` for now.

The `CheckRackScale()` is inside `pkg/reconcilation/reconcile_rack.go`. The key idea is if the current number of replicas in a rack is less than the desired number, it triggers a scale-up by updating the StatefulSet.

The function iterates through each rack (which was previously calculated in `CalculateRackInformation`):

```go
for idx := range rc.desiredRackInformation {
    rackInfo := rc.desiredRackInformation[idx]
    statefulSet := rc.statefulSets[idx]

    desiredNodeCount := int32(rackInfo.NodeCount)
    maxReplicas := *statefulSet.Spec.Replicas
```

For each rack, it gets the:

- Desired node count (from the rack information calculated earlier)
- Current replica count in the StatefulSet

Then checks if scaling is needed:

```go
if maxReplicas < desiredNodeCount {
```

If scaling is needed (`current replicas < desired nodes`), it sets the appropriate condition:

```go
if err := rc.setConditionStatus(api.DatacenterScalingUp, corev1.ConditionTrue); err != nil
```

Which is the whole point of this section.

### 3. Increment replicas by 2

We still inside `CheckRackScale()`. The actual StatefulSet update happens in the `UpdateRackNodeCount` function:

This function (beside logging):

```go
patch := client.MergeFrom(statefulSet.DeepCopy())
statefulSet.Spec.Replicas = &newNodeCount
err := rc.Client.Patch(rc.Ctx, statefulSet, patch)
```

- Creates a patch of the StatefulSet
- Updates the replica count to the desired node count
- Sends the PATCH request to Kubernetes. If successful, Kubernetes updates the StatefulSet, and its controller starts creating the additional pods.

### 4. Update Event, Create Pods In Parallel

These operations are outside the scope of this opeartor.

### 5. Check Node Manage API running

After updating the StatefulSet, the main reconciliation loop calls `ReconcileAllRacks` repeatedly. If you still remember, it still contains multiple checks after the `CheckRackScale()`, including:

- `CheckPodsReady()`
- `CheckCassandraNodeStatuses()`

The `CheckPodsReady()` function is responsible for `requeue reconcile event until Node Management API running` in the documentation. It tries to start Cassandra nodes through various helper functions

The Node Management API (NMAPI) is a lightweight HTTP server that runs inside every Cassandra pod and exposes endpoints for:

- Bootstrapping (/lifecycle/start)
- Refreshing seeds
- Updating configs
- Health checks

The `cass-operator` interacts with each Cassandra node through this API. It checks if NMAPI is reachable on a pod. If it isn’t, the reconcile returns something like:

```go
return result.RequeueSoon(2)
```

This causes the controller to retry in 2 seconds.

Additional informaton:

The key function for checking if the Management API is running is `isMgmtApiRunning` (`CheckPodsReady` -> (`startBootstrappedNodes` | `startOneNodePerRack` | `startAllNodes`) -> `startNode` -> `isMgmtApiRunning`):

This function:

- Checks if the "cassandra" container is running
- If it's running, ensures it has been running for at least 10 seconds (to give the Management API time to initialize)
- Returns true only if the container has been running for more than 10 seconds

### 6. Start cassandra

Inside the `startNode` function, when the Management API is confirmed to be running (passing the `isMgmtApiRunning` check), the operator sends a request to the Management API to start the Cassandra node:

```go
func (rc *ReconciliationContext) startNode(pod *corev1.Pod, labelSeedBeforeStart bool, endpointData httphelper.CassMetadataEndpoints) (bool, error) {

    ...
    if !isServerReady(pod) {
        if isServerReadyToStart(pod) && isMgmtApiRunning(pod) { // -> check here
            ...

            // Start Cassandra
            if err := rc.startCassandra(endpointData, pod); err != nil { // -> start cassandra here
                return true, err
            }
        }
        return true, nil
    }
    return false, nil
}
```

After these process is done, all the remaining events are triggered, but outside the scope of the operator including:

- update event (pod ready) (from Pod to StatefulSet Controller)
- increment ready replicas (from StatefulSet Controller to cluster StatefulSet)
- update event (from cluster StatetfulSet to the operator)

Then the operator run the `Reconcile` function again, and the process continues until all nodes are up and running. After that it sets the `ScalingUp` condition to `False`.

### How New Nodes Join the Cassandra Ring

The primary mechanism for new nodes to join the cluster is through seed node discovery. It happen when StatefulSet controller creates the pods.

First, In `pkg/serconfig/configgen.go`, the operator defines seed nodes in the Cassandra configuration:

```go
seeds := []string{dc.GetSeedServiceName(), dc.GetAdditionalSeedsServiceName()}
```

This adds two special Kubernetes service names to the seed list in cassandra.yaml:

- `<cluster-name>-seed-service` - Points to pods labeled as seeds
- `<cluster-name>-<dc-name>`-additional-seed-service - Points to any additional seeds

Before the configuration, the operator creates Kubernetes services that target seed pods:

`reconcile_services.go` -> `newSeedServiceForCassandraDatacenter(dc)` -> `GetSeedServiceName`

This service is headless (no ClusterIP) and allows DNS discovery of seed nodes. When a pod does a DNS lookup for `<cluster-name>-seed-service`, Kubernetes DNS returns the IPs of all seed pod

When new nodes are added to the cluster, the operator ensures existing nodes know about them:

```go
func (rc \*ReconciliationContext) refreshSeeds() error {
// ...
startedPods := FilterPodListByCassNodeState(rc.clusterPods, stateStarted)

    for _, pod := range startedPods {
        if err := rc.NodeMgmtClient.CallReloadSeedsEndpoint(pod); err != nil {
            return err
        }
    }
    // ...

}
```

When starting Cassandra on a new node, the Management API handles the details of:

- Reading the seeds from `cassandra.yaml`
- Initializing gossip communication with seed nodes
- Joining the Cassandra ring
- Streaming data from existing nodes

### How the operator distributes Cassandra racks across different availability zones?
