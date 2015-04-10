package control

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/intelsdilabs/pulse/control/plugin/cpolicy"
	"github.com/intelsdilabs/pulse/core"
	"github.com/intelsdilabs/pulse/core/cdata"
	"github.com/intelsdilabs/pulse/core/ctypes"
	"github.com/intelsdilabs/pulse/core/mttrie"
)

var (
	errMetricNotFound   = errors.New("metric not found")
	errNegativeSubCount = errors.New("subscription count cannot be < 0")
)

type metricType struct {
	Plugin             *loadedPlugin
	namespace          []string
	lastAdvertisedTime time.Time
	subscriptions      int
	policy             processesConfigData
	config             *cdata.ConfigDataNode
}

type processesConfigData interface {
	Process(map[string]ctypes.ConfigValue) (*map[string]ctypes.ConfigValue, *cpolicy.ProcessingErrors)
}

func newMetricType(ns []string, last time.Time, plugin *loadedPlugin) *metricType {
	return &metricType{
		Plugin: plugin,

		namespace:          ns,
		lastAdvertisedTime: last,
	}
}

func (m *metricType) Namespace() []string {
	return m.namespace
}

func (m *metricType) LastAdvertisedTime() time.Time {
	return m.lastAdvertisedTime
}

func (m *metricType) Subscribe() {
	m.subscriptions++
}

func (m *metricType) Unsubscribe() error {
	if m.subscriptions == 0 {
		return errNegativeSubCount
	}
	m.subscriptions--
	return nil
}

func (m *metricType) SubscriptionCount() int {
	return m.subscriptions
}

func (m *metricType) Version() int {
	if m.Plugin == nil {
		return -1
	}
	return m.Plugin.Version()
}

func (m *metricType) Config() *cdata.ConfigDataNode {
	return nil
}

type metricCatalog struct {
	tree        *mttrie.MTTrie
	mutex       *sync.Mutex
	keys        []string
	currentIter int
}

func newMetricCatalog() *metricCatalog {
	var k []string
	return &metricCatalog{
		tree:        mttrie.New(),
		mutex:       &sync.Mutex{},
		currentIter: 0,
		keys:        k,
	}
}

func (m *metricCatalog) AddLoadedMetricType(lp *loadedPlugin, mt core.MetricType) {
	if lp.ConfigPolicyTree == nil {
		panic("NO")
	}

	newMt := metricType{
		Plugin:             lp,
		namespace:          mt.Namespace(),
		lastAdvertisedTime: mt.LastAdvertisedTime(),
		policy:             lp.ConfigPolicyTree.Get(mt.Namespace()),
	}
	m.Add(&newMt)
}

// Add adds a metricType
func (mc *metricCatalog) Add(m *metricType) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	key := getMetricKey(m.Namespace())
	mc.keys = appendIfMissing(mc.keys, key)

	mc.tree.Add(m.Namespace(), m)
}

// Get retrieves a loadedPlugin given a namespace and version.
// If provided a version of -1 the latest plugin will be returned.
func (mc *metricCatalog) Get(ns []string, version int) (*metricType, error) {
	mc.Lock()
	defer mc.Unlock()
	//key := getMetricKey(ns)
	return mc.get(ns, version)
}

// Fetch transactionally retrieves all loadedPlugins
func (mc *metricCatalog) Fetch(ns []string) ([]core.MetricType, error) {
	mc.Lock()
	defer mc.Unlock()

	mtsi, err := mc.tree.Fetch(ns)
	if err != nil {
		return nil, err
	}
	return mtsi, nil
}

// used to lock the plugin table externally,
// when iterating in unsafe scenarios
func (mc *metricCatalog) Lock() {
	mc.mutex.Lock()
}

func (mc *metricCatalog) Unlock() {
	mc.mutex.Unlock()
}

func (mc *metricCatalog) Remove(ns []string) {
	mc.mutex.Lock()
	mc.tree.Remove(ns)
	mc.mutex.Unlock()
}

// Item returns the current metricType in the collection.  The method Next()
// provides the  means to move the iterator forward.
func (mc *metricCatalog) Item() (string, []*metricType) {
	key := mc.keys[mc.currentIter-1]
	ns := strings.Split(key, ".")
	mtsi, _ := mc.tree.Get(ns)
	var mts []*metricType
	for _, mt := range mtsi {
		mts = append(mts, mt.(*metricType))
	}
	return key, mts
}

// Next returns true until the "end" of the collection is reached.  When
// the end of the collection is reached the iterator is reset back to the
// head of the collection.
func (mc *metricCatalog) Next() bool {
	mc.currentIter++
	if mc.currentIter > len(mc.keys) {
		mc.currentIter = 0
		return false
	}
	return true
}

// Subscribe atomically increments a metric's subscription count in the table.
func (mc *metricCatalog) Subscribe(ns []string, version int) error {
	mc.Lock()
	defer mc.Unlock()

	m, err := mc.get(ns, version)
	if err != nil {
		return err
	}

	m.Subscribe()
	return nil
}

// Unsubscribe atomically decrements a metric's count in the table
func (mc *metricCatalog) Unsubscribe(ns []string, version int) error {
	mc.Lock()
	defer mc.Unlock()

	m, err := mc.get(ns, version)
	if err != nil {
		return err
	}

	return m.Unsubscribe()
}

func (mc *metricCatalog) GetPlugin(mns []string, ver int) (*loadedPlugin, error) {
	m, err := mc.Get(mns, ver)
	if err != nil {
		return nil, err
	}
	return m.Plugin, nil
}

func (mc *metricCatalog) get(ns []string, ver int) (*metricType, error) {
	mts, err := mc.tree.Get(ns)
	if err != nil {
		return nil, err
	}

	if len(mts) > 1 {
		// a version IS given
		if ver >= 0 {
			l, err := getVersion(mts, ver)
			if err != nil {
				return nil, err
			}
			//TODO Can we avoid this type assert?
			return l.(*metricType), nil
		}

		// multiple versions but -1 was given for the version
		// meaning get the latest
		return getLatest(mts).(*metricType), nil
	}

	//only one version so return it
	return mts[0].(*metricType), nil

}

func getMetricKey(metric []string) string {
	return strings.Join(metric, ".")
}

func getLatest(c []core.MetricType) core.MetricType {
	cur := c[0]
	for _, mt := range c {
		if mt.Version() > cur.Version() {
			cur = mt
		}

	}
	return cur
}

func appendIfMissing(keys []string, ns string) []string {
	for _, key := range keys {
		if ns == key {
			return keys
		}
	}
	return append(keys, ns)
}

//TODO ? The trie could expose a GetByVersion eliminating the need for this
func getVersion(c []core.MetricType, ver int) (core.MetricType, error) {
	for _, m := range c {
		if m.(*metricType).Plugin.Version() == ver {
			return m, nil
		}
	}
	return nil, errMetricNotFound
}
