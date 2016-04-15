# snapd

## build & run
```
go build -o ./build/bin/snapd github.com/intelsdi-x/snap
./build/bin/snapd --plugin-trust=0 --log-level=1 --config=serenity2/tasks/mesos-global-conf.json
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
go install github.com/intelsdi-x/snap/serenity2/p3-invert
go install github.com/intelsdi-x/snap/serenity2/p4-new
go install github.com/intelsdi-x/snap/serenity2/o1
go install github.com/intelsdi-x/snap/serenity2/o2
``` 


## load 
```
./build/bin/snapctl plugin load `which i1`
./build/bin/snapctl plugin load `which i2`
./build/bin/snapctl plugin load `which p1`
./build/bin/snapctl plugin load `which p2`
./build/bin/snapctl plugin load `which p3-invert`
./build/bin/snapctl plugin load `which p4-new`
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

### chain
```
# just passing
./build/bin/snapctl task create -t serenity2/tasks/chain.yaml
# invert
./build/bin/snapctl task create -t serenity2/tasks/chain-invert.yaml
# new
./build/bin/snapctl task create -t serenity2/tasks/chain-new.yaml
# long (pass-new-invert-pass)
./build/bin/snapctl task create -t serenity2/tasks/chain-long.yaml
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


## Experimenting

### mutilate build & compile

(cd serenity2/heracles/mutilate/; scons)

./serenity2/heracles/mutilate/mutilate --version

### memcache build & compile
(cd serenity2/heracles/memcached/; ./autogen.sh && ./configure && make)

./serenity2/heracles/memcached/memcached -V

### aliases
alias memcached=./serenity2/heracles/memcached/memcached mutilate=./serenity2/heracles/mutilate/mutilate

### search
mutilate --server 127.0.0.1 --search 95:1000 --time 1 --save tmp/search.log

### scan
mutilate --server 127.0.0.1 --scan 0:80000:10000 --time 1 --save tmp/scan.log

### cgroups prepare (as root)
sudo -Es

mkdir /sys/fs/cgroup/cpuset/prod
mkdir /sys/fs/cgroup/cpuset/be

#### clean
cgdelete -g cpuset:/prod 
cgdelete -g cpuset:/be

### set cpus/mems
echo 0 > /sys/fs/cgroup/cpuset/prod/cpuset.mems
echo 0 > /sys/fs/cgroup/cpuset/prod/cpuset.cpus

### memcache
cgexec -g cpuset:/prod ./serenity2/heracles/memcached/memcached


# long chain output
```
---------------- P1 just pass --------------------------
2016/04/13 16:39:30 2016/04/13 16:39:30 processor:Process called
2016/04/13 16:39:30 2016/04/13 16:39:30 PROCESSOR-DUMP: 2016-04-13 16:39:30.987993735 +0200 CEST|[serenity2 input metric1]|42

---------------- P4 new -------------------
2016/04/13 16:39:30 2016/04/13 16:39:30 processor:Process called
2016/04/13 16:39:30 2016/04/13 16:39:30 PROCESS-METRICS: (func([]plugin.PluginMetricType) []plugin.PluginMetricType)(0x4011a0)
2016/04/13 16:39:30 2016/04/13 16:39:30 new: 43
2016/04/13 16:39:30 2016/04/13 16:39:30 PROCESSOR-DUMP: 2016-04-13 16:39:30.987993735 +0200 CEST|[serenity2 input metric1]|42
2016/04/13 16:39:30 2016/04/13 16:39:30 PROCESSOR-DUMP: 2016-04-13 16:39:30.989330084 +0200 CEST|[serenity2 process metric-p4-new]|43

---------------- P3 invert -----------------
2016/04/13 16:39:30 2016/04/13 16:39:30 processor:Process called
2016/04/13 16:39:30 2016/04/13 16:39:30 PROCESS-METRICS: (func([]plugin.PluginMetricType) []plugin.PluginMetricType)(0x4011a0)
2016/04/13 16:39:30 2016/04/13 16:39:30 inverted: -42
2016/04/13 16:39:30 2016/04/13 16:39:30 inverted: -43
2016/04/13 16:39:30 2016/04/13 16:39:30 PROCESSOR-DUMP: 2016-04-13 16:39:30.987993735 +0200 CEST|[serenity2 input metric1]|-42
2016/04/13 16:39:30 2016/04/13 16:39:30 PROCESSOR-DUMP: 2016-04-13 16:39:30.989330084 +0200 CEST|[serenity2 process metric-p4-new]|-43

---------------- P2 pass again --------------------
2016/04/13 16:39:30 2016/04/13 16:39:30 processor:Process called
2016/04/13 16:39:30 2016/04/13 16:39:30 PROCESSOR-DUMP: 2016-04-13 16:39:30.987993735 +0200 CEST|[serenity2 input metric1]|-42
2016/04/13 16:39:30 2016/04/13 16:39:30 PROCESSOR-DUMP: 2016-04-13 16:39:30.989330084 +0200 CEST|[serenity2 process metric-p4-new]|-43

--------------- O2 publish --------------------
2016/04/13 16:39:30 2016/04/13 16:39:30 output:Publish called
2016/04/13 16:39:30 2016/04/13 16:39:30 PUBLISHER-DUMP: 2016-04-13 16:39:30.987993735 +0200 CEST|[serenity2 input metric1]|-42
2016/04/13 16:39:30 2016/04/13 16:39:30 PUBLISHER-DUMP: 2016-04-13 16:39:30.989330084 +0200 CEST|[serenity2 process metric-p4-new]|-43
```

### real collectors

#### Mesos agent

##### deps
go get github.com/intelsdi-x/snap-plugin-utilities/ns

##### build
go get -v github.com/intelsdi-x/snap-plugin-collector-mesos
(
cd $GOPATH/src/github.com/intelsdi-x/snap-plugin-collector-mesos
git fetch origin pull/1/head:marcin-krolik/kromar-mesos-wip
git co marcin-krolik:kromar-mesos-wip
go build
ls $GOPATH/src/github.com/intelsdi-x/snap-plugin-collector-mesos/snap-plugin-collector-mesos
)

##### load 
./build/bin/snapctl plugin load `which snap-plugin-collector-mesos`

#### task
./build/bin/snapctl task create -t serenity2/tasks/mesos.yaml


#### example output
http://127.0.0.1:5051/monitor/statistics


>> task watch
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0008/cpus_limit                   1.1                     2016-04-15 15:36:52.675641984 +0200 CEST        gklab-126-081
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0008/cpus_system_time_secs        0.03                    2016-04-15 15:36:52.675641984 +0200 CEST        gklab-126-081
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0008/cpus_user_time_secs          14.32                   2016-04-15 15:36:52.675641984 +0200 CEST        gklab-126-081
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0008/mem_limit_bytes              1.6777216e+08           2016-04-15 15:36:52.675641984 +0200 CEST        gklab-126-081
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0008/perf/duration                0                       2016-04-15 15:36:52.675641984 +0200 CEST        gklab-126-081
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0008/perf/timestamp               1.46072739835631e+09    2016-04-15 15:36:52.675641984 +0200 CEST        gklab-126-081
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0008/timestamp                    1.46072741267457e+09    2016-04-15 15:36:52.675641984 +0200 CEST        gklab-126-081
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0008/cpus_limit                   1.1                     2016-04-15 15:37:22.675926443 +0200 CEST        gklab-126-081
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0008/cpus_system_time_secs        0.06                    2016-04-15 15:37:22.675926443 +0200 CEST        gklab-126-081
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0008/cpus_user_time_secs          44.36                   2016-04-15 15:37:22.675926443 +0200 CEST        gklab-126-081
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0008/mem_limit_bytes              1.6777216e+08           2016-04-15 15:37:22.675926443 +0200 CEST        gklab-126-081
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0008/perf/duration                1                       2016-04-15 15:37:22.675926443 +0200 CEST        gklab-126-081
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0008/perf/instructions            7.645007664e+09         2016-04-15 15:37:22.675926443 +0200 CEST        gklab-126-081
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0008/perf/timestamp               1.46072743537553e+09    2016-04-15 15:37:22.675926443 +0200 CEST        gklab-126-081
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0008/timestamp                    1.46072744267483e+09    2016-04-15 15:37:22.675926443 +0200 CEST        gklab-126-081

>> snapctl metric list


# other collectors

go get github.com/intelsdi-x/snap-plugin-collector-pcm
go get github.com/intelsdi-x/snap-plugin-collector-perfevents
go install github.com/intelsdi-x/snap-plugin-collector-pcm
go install github.com/intelsdi-x/snap-plugin-collector-perfevents

./build/bin/snapctl plugin load `which snap-plugin-collector-pcm`
./build/bin/snapctl plugin load `which snap-plugin-collector-perfevents`

./build/bin/snapctl task create -t serenity2/tasks/pcm.yaml

```
NAMESPACE                                                                                        VERSIONS
/intel/mesos/agent/*/foo                                                                         1
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0007/cpus_limit                   1
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0007/cpus_system_time_secs        1
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0007/cpus_user_time_secs          1
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0007/mem_limit_bytes              1
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0007/perf/duration                1
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0007/perf/instructions            1
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0007/perf/timestamp               1
/intel/mesos/agent/stress/b2cf6b8b-e302-4afa-bc20-e93c7fce5e23-0007/timestamp                    1

/intel/pcm/ACYC                  321             2016-04-15 16:33:20.813281473 +0200 CEST        gklab-126-081
/intel/pcm/AFREQ                 0.0915          2016-04-15 16:33:20.813278144 +0200 CEST        gklab-126-081
/intel/pcm/C0res%                12.8            2016-04-15 16:33:20.813283589 +0200 CEST        gklab-126-081
/intel/pcm/C1res%                18.2            2016-04-15 16:33:20.813283856 +0200 CEST        gklab-126-081
/intel/pcm/C2res%                0               2016-04-15 16:33:20.813285212 +0200 CEST        gklab-126-081
/intel/pcm/C3res%                0               2016-04-15 16:33:20.813284189 +0200 CEST        gklab-126-081
/intel/pcm/C6res%                0               2016-04-15 16:33:20.813284469 +0200 CEST        gklab-126-081
/intel/pcm/C7res%                0               2016-04-15 16:33:20.813284739 +0200 CEST        gklab-126-081
/intel/pcm/EXEC                  0.0249          2016-04-15 16:33:20.813276765 +0200 CEST        gklab-126-081
/intel/pcm/FREQ                  0.0117          2016-04-15 16:33:20.81327771 +0200 CEST         gklab-126-081
/intel/pcm/INST                  682             2016-04-15 16:33:20.813281185 +0200 CEST        gklab-126-081
/intel/pcm/INSTnom               0.0498          2016-04-15 16:33:20.813282814 +0200 CEST        gklab-126-081
/intel/pcm/INSTnom%              1.24            2016-04-15 16:33:20.813283214 +0200 CEST        gklab-126-081
/intel/pcm/IPC                   2.12            2016-04-15 16:33:20.813277354 +0200 CEST        gklab-126-081
/intel/pcm/L2CLK                 0.0427          2016-04-15 16:33:20.813280189 +0200 CEST        gklab-126-081
/intel/pcm/L2HIT                 0.233           2016-04-15 16:33:20.81327954 +0200 CEST         gklab-126-081
/intel/pcm/L2MISS                0.602           2016-04-15 16:33:20.813278971 +0200 CEST        gklab-126-081
/intel/pcm/L3CLK                 0.158           2016-04-15 16:33:20.813279862 +0200 CEST        gklab-126-081
/intel/pcm/L3HIT                 0.532           2016-04-15 16:33:20.813279254 +0200 CEST        gklab-126-081
/intel/pcm/L3MISS                0.282           2016-04-15 16:33:20.813278602 +0200 CEST        gklab-126-081
/intel/pcm/PhysIPC               4.25            2016-04-15 16:33:20.8132822 +0200 CEST          gklab-126-081
/intel/pcm/PhysIPC%              106             2016-04-15 16:33:20.813282543 +0200 CEST        gklab-126-081
/intel/pcm/Proc_Energy_Joules    25.3            2016-04-15 16:33:20.81328561 +0200 CEST         gklab-126-081
/intel/pcm/READ                  0.0234          2016-04-15 16:33:20.81328058 +0200 CEST         gklab-126-081
/intel/pcm/SKT0                  25.3            2016-04-15 16:33:20.813286316 +0200 CEST        gklab-126-081
/intel/pcm/TEMP                  51              2016-04-15 16:33:20.813285921 +0200 CEST        gklab-126-081
/intel/pcm/TIME_ticks            3430            2016-04-15 16:33:20.81328189 +0200 CEST         gklab-126-081
/intel/pcm/WRITE                 0.0161          2016-04-15 16:33:20.813280863 +0200 CEST        gklab-126-081
```
