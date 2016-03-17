# snapd

## build & run
```
go build -o ./build/bin/snapd github.com/intelsdi-x/snap
./build/bin/snapd --plugin-trust=0 --log-level=1
```

### snapctl
```
go build -o ./build/bin/snapctl github.com/intelsdi-x/snap/cmd/snapctl
```

# plugins
i - input - collector plugin
p - procssor plugin
o - output - publisher plugin

## build
```
go install github.com/intelsdi-x/snap/serenity2/i1
go install github.com/intelsdi-x/snap/serenity2/i2
go install github.com/intelsdi-x/snap/serenity2/p1
go install github.com/intelsdi-x/snap/serenity2/p2
go install github.com/intelsdi-x/snap/serenity2/o1
go install github.com/intelsdi-x/snap/serenity2/o2
``` 


## load 
```
./build/bin/snapctl plugin load `which i1`
./build/bin/snapctl plugin load `which i2`
./build/bin/snapctl plugin load `which p1`
./build/bin/snapctl plugin load `which p2`
./build/bin/snapctl plugin load `which o1`
./build/bin/snapctl plugin load `which o2`
```


## reload just one plugin eg. p1
```
./build/bin/snapctl plugin unload processor:p1:1
./build/bin/snapctl plugin load `which p1`
```


## list
```
./build/bin/snapctl plugin list 
```

# metrics
```
./build/bin/snapctl metric list 
```

# tasks

## create

### simple tasks
```
./build/bin/snapctl task create -t serenity2/tasks/simple.yaml
./build/bin/snapctl task create -t serenity2/tasks/2i_1p_2o.yaml
```

### complex
```
./build/bin/snapctl task create -t serenity2/tasks/complex.yaml
```

## list 
```
./build/bin/snapctl task list 
TASK_ID=`./build/bin/snapctl task list | awk '/Task-/ {print $1}'`
```

## watch
```
./build/bin/snapctl task watch $TASK_ID
```

## stop & remove
```
./build/bin/snapctl task stop $TASK_ID
./build/bin/snapctl task remove $TASK_ID
```



# HTTP API
```
curl -L http://localhost:8181/v1/plugins
curl -L http://localhost:8181/v1/metrics
curl -L http://localhost:8181/v1/tasks
```

# docs

https://github.com/intelsdi-x/snap/blob/master/docs/TASKS.md
https://github.com/intelsdi-x/snap/blob/master/docs/REST_API.md#task-api

## simultanous & synchornization of running plugins (second layer)

"Ensures scheduler jobs submissions are sent concurrently for a workflow."
https://github.com/intelsdi-x/snap/pull/743

if collector publsishe * then it on him
else framework take care

# plugins logs

## just from core framework (a.k.a. SessionState.logger)
```
tail -n 0 -F /tmp/p*.log -F /tmp/i*.log -F /tmp/o*.log
```

## just errors (log.Println)
```
tail -n 0 -F /tmp/i*.stderr -F /tmp/p*.stderr -F /tmp/o*.stderr
```

## all
```
tail -n 0 -F /tmp/p*.log -F /tmp/i*.log -F /tmp/o*.log -F /tmp/i*.stderr -F /tmp/p*.stderr -F /tmp/o*.stderr
```

## logging tips

### golang
log.Println -> stderr
fmt.Println -> stdout

### snap
```
plugin.SessionState.Logger (s.Logger()) -> /tmp/PLUGIN_NAME.loger
plugin (stderr) -> /tmp/PLUGIN_NAME.stderr (thx to logStdErr goroutine)
plugin (stdout) -> ???
plugin.execution.execLogger -> snapd.log (DAEMON! or stderr)
```


### Plugin options

```
// AcceptedContentTypes are types accepted by this plugin in priority order.
// snap.* means any snap type.
AcceptedContentTypes []string

// ReturnedContentTypes are content types returned in priority order.
// This is only applicable on processors.
ReturnedContentTypes []string

// ConcurrencyCount is the max number concurrent calls the plugin may take.
// If there are 5 tasks using the plugin and concurrency count is 2 there
// will be 3 plugins running.
ConcurrencyCount int 
    // default to 1

// Exclusive results in a single instance of the plugin running regardless
// the number of tasks using the plugin.
Exclusive bool
    // Checking if plugin is exclusive
    // (only one instance should be running).
    if a.Exclusive() {
        p.max = 1
    }

// Unsecure results in unencrypted communication with this plugin.
Unsecure bool

// CacheTTL will override the default cache TTL for the provided plugin.
CacheTTL time.Duration

// RoutingStrategy will override the routing strategy this plugin requires.
// The default routing strategy round-robin.
RoutingStrategy RoutingStrategyType
    https://github.com/intelsdi-x/snap/issues/539
    // Set the routing and caching strategy
    // DefaultRouting is a least recently used strategy.
    DefaultRouting RoutingStrategyType = iota
    // StickyRouting is a one-to-one strategy.
    // Using this strategy a tasks requests are sent to the same running instance of a plugin.
    // sticky provides a stragey that ... concurrency count is 1
    StickyRouting
    // ConfigRouting is routing to plugins based on the config provided to the plugin.
    // Using this strategy enables a running database plugin that has the same connection info between
    // two tasks to be shared.
    ConfigRouting
    // NOT IMPLEMENTED !!!!!! ????
```

# plugins states

## after loading
loadedPlugin(pluginMeta)  managed by pluginManager(managesPlugin) 

## after starting
availablePlugin(exectuablePlugin(cmd)) managed by pluginRunner(runsPlugins)

## under the hood

```
pluginControl (plugin/control.go)

    # state
    RunningPlugins []plugin.ExecutablePlugin

    ## behavior
    pluginManager (loading/unloading plugins, control/plugin_manager.go)
    pluginRunner (running/stopping plugins, control/runner.go)
	metricCatalog  catalogsMetrics(manages metrics, mapping from namespace -> loadedPlugin, control/metrics.go)


Scheduler uses pluginControl as metricManager
and tasks use metricManager to gather the metrics !

```

# Errors/Limitations

## 1. Unsupported parent job type/unsupported content type/unsupported type

log:

```
FATA[0530] unsupported parent job type                   _module=scheduler-job block=run content-type=snap.gob job-type=processor parent-job-type=2 plugin-config=map[] plugin-name=p2 plugin-version=-1
```

### schedule/job package

coreJob
    AddErrors()
    
collectorJob
    coreJob
    Run()
        .metrics = .collector.CollectMetrics(requestedMetricTypes)

processJob
    coreJob
    Run()
        if .parrent == "collector":
            if .contentType == snap.gob:
                content = encode(p.metrics)
                .content = .processor.ProcessMetrics(content)
            else: 
                panic("unsupported content type")
        else:
            panic("unsupported parent job type")

publisherJob
    coreJob
    Run()
        if .parent == "collector":
            if .contentType == snap.gob:
                content = encode(p.metrics)
                .publisher.PublisherMetrics(content)
            else:
                panic("unsupported content type")
        elif .parent == "processor"
            if .contentType == snap.gob:
                .publisher.PublishMertics(.parent.content)
            else:
                "nothing happens!"
        else:
            panic("unsupported parent job type")
            



## 2. Metric catalog quering

Cannot query of metrics with general query
```
Metric not found: /serenity2/input/*
```
example task:

./build/bin/snapctl task create -t serenity2/tasks/query.yaml


## 3. Processors and publishers clones
```
./build/bin/snapctl task create -t serenity2/tasks/processor_clones.yaml
./build/bin/snapctl task create -t serenity2/tasks/output_clones.yaml
```

