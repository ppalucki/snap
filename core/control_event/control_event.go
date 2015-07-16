package control_event

const (
	AvailablePluginDead   = "Control.AvailablePluginDead"
	PluginLoaded          = "Control.PluginLoaded"
	PluginUnloaded        = "Control.PluginUnloaded"
	PluginsSwapped        = "Control.PluginsSwapped"
	PublisherSubscribed   = "Control.PublisherSubscribed"
	PublisherUnsubscribed = "Control.PublisherUnsubscribed"
	ProcessorSubscribed   = "Control.ProcessorSubscribed"
	ProcessorUnsubscribed = "Control.ProcessorUnsubscribed"
	MetricSubscribed      = "Control.MetricSubscribed"
	MetricUnsubscribed    = "Control.MetricUnsubscribed"
	HealthCheckFailed     = "Control.PluginHealthCheckFailed"
)

type LoadPluginEvent struct {
	Name    string
	Version int
	Type    int
}

func (e LoadPluginEvent) Namespace() string {
	return PluginLoaded
}

type UnloadPluginEvent struct {
	Name    string
	Version int
	Type    int
}

func (e UnloadPluginEvent) Namespace() string {
	return PluginUnloaded
}

type DeadAvailablePluginEvent struct {
	Name    string
	Version int
	Type    int
	Key     string
	Id      int
	String  string
}

func (e *DeadAvailablePluginEvent) Namespace() string {
	return AvailablePluginDead
}

type SwapPluginsEvent struct {
	LoadedPluginName      string
	LoadedPluginVersion   int
	UnloadedPluginName    string
	UnloadedPluginVersion int
	PluginType            int
}

func (s SwapPluginsEvent) Namespace() string {
	return PluginsSwapped
}

type PublisherSubscriptionEvent struct {
	PluginName    string
	PluginVersion int
}

func (se PublisherSubscriptionEvent) Namespace() string {
	return PublisherSubscribed
}

type PublisherUnsubscriptionEvent struct {
	PluginName    string
	PluginVersion int
}

func (puue PublisherUnsubscriptionEvent) Namespace() string {
	return PublisherUnsubscribed
}

type ProcessorSubscriptionEvent struct {
	PluginName    string
	PluginVersion int
}

func (se ProcessorSubscriptionEvent) Namespace() string {
	return ProcessorSubscribed
}

type ProcessorUnsubscriptionEvent struct {
	PluginName    string
	PluginVersion int
}

func (prue ProcessorUnsubscriptionEvent) Namespace() string {
	return ProcessorUnsubscribed
}

type MetricSubscriptionEvent struct {
	MetricNamespace []string
	Version         int
}

func (se MetricSubscriptionEvent) Namespace() string {
	return MetricSubscribed
}

type MetricUnsubscriptionEvent struct {
	MetricNamespace []string
}

func (ue MetricUnsubscriptionEvent) Namespace() string {
	return MetricUnsubscribed
}

type HealthCheckFailedEvent struct {
	Name    string
	Version int
	Type    int
}

func (hfe HealthCheckFailedEvent) Namespace() string {
	return HealthCheckFailed
}
