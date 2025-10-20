package apollo

import (
	"container/list"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/apolloconfig/agollo/v4/agcache"
	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/stretchr/testify/assert"
)

// mockApolloClient mocks the Apollo client for testing
type mockApolloClient struct {
	configCache *mockCache
	listeners   *list.List
}

func newMockApolloClient() *mockApolloClient {
	return &mockApolloClient{
		configCache: newMockCache(),
		listeners:   list.New(),
	}
}

func (m *mockApolloClient) GetConfigCache(namespace string) agcache.CacheInterface {
	return m.configCache
}

func (m *mockApolloClient) GetDefaultConfigCache() agcache.CacheInterface {
	return m.configCache
}

func (m *mockApolloClient) GetApolloConfigCache() agcache.CacheInterface {
	return m.configCache
}

func (m *mockApolloClient) GetValue(key string) string {
	val, _ := m.configCache.Get(key)
	if val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

func (m *mockApolloClient) GetStringValue(key string, defaultValue string) string {
	val := m.GetValue(key)
	if val == "" {
		return defaultValue
	}
	return val
}

func (m *mockApolloClient) GetIntValue(key string, defaultValue int) int {
	return defaultValue
}

func (m *mockApolloClient) GetFloatValue(key string, defaultValue float64) float64 {
	return defaultValue
}

func (m *mockApolloClient) GetBoolValue(key string, defaultValue bool) bool {
	return defaultValue
}

func (m *mockApolloClient) GetConfig(namespace string) *storage.Config {
	return nil
}

func (m *mockApolloClient) GetConfigAndInit(namespace string) *storage.Config {
	return nil
}

func (m *mockApolloClient) GetStringSliceValue(key string, defaultValue []string) []string {
	return defaultValue
}

func (m *mockApolloClient) GetIntSliceValue(key string, defaultValue []int) []int {
	return defaultValue
}

func (m *mockApolloClient) AddChangeListener(listener storage.ChangeListener) {
	m.listeners.PushBack(listener)
}

func (m *mockApolloClient) RemoveChangeListener(listener storage.ChangeListener) {
	// Not implemented for testing
}

func (m *mockApolloClient) GetChangeListeners() *list.List {
	return m.listeners
}

func (m *mockApolloClient) UseEventDispatch() {
	// Not implemented for testing
}

func (m *mockApolloClient) Close() {
	// Not implemented for testing
}

// mockCache implements agcache.CacheInterface
type mockCache struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

func newMockCache() *mockCache {
	return &mockCache{
		data: make(map[string]interface{}),
	}
}

func (m *mockCache) Set(key string, value interface{}, expireSeconds int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
	return nil
}

func (m *mockCache) EntryCount() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int64(len(m.data))
}

func (m *mockCache) Get(key string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, exists := m.data[key]
	if !exists {
		return nil, errors.New("key not found")
	}
	return val, nil
}

func (m *mockCache) Del(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, exists := m.data[key]
	if exists {
		delete(m.data, key)
	}
	return exists
}

func (m *mockCache) Range(f func(key, value interface{}) bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for k, v := range m.data {
		if !f(k, v) {
			break
		}
	}
}

func (m *mockCache) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]interface{})
}

// TestApolloSubscriber_LoadValue_JSON tests loading JSON configuration
func TestApolloSubscriber_LoadValue_JSON(t *testing.T) {
	client := newMockApolloClient()
	client.configCache.Set("name", "test-app", 0)
	client.configCache.Set("version", "1.0.0", 0)
	client.configCache.Set("timeout", float64(30), 0)

	sub := &apolloSubscriber{
		client: client,
		conf: ApolloConf{
			NamespaceName: "application.json",
			Format:        "json",
		},
	}

	err := sub.loadValue()
	assert.NoError(t, err)

	value, err := sub.Value()
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(value), &result)
	assert.NoError(t, err)
	assert.Equal(t, "test-app", result["name"])
	assert.Equal(t, "1.0.0", result["version"])
	assert.Equal(t, float64(30), result["timeout"])
}

// TestApolloSubscriber_LoadValue_SpecificKey tests loading a specific key
func TestApolloSubscriber_LoadValue_SpecificKey(t *testing.T) {
	client := newMockApolloClient()
	client.configCache.Set("database.url", "mysql://localhost:3306/test", 0)
	client.configCache.Set("database.user", "root", 0)

	sub := &apolloSubscriber{
		client: client,
		conf: ApolloConf{
			NamespaceName: "application",
			Key:           "database.url",
		},
	}

	err := sub.loadValue()
	assert.NoError(t, err)

	value, err := sub.Value()
	assert.NoError(t, err)
	assert.Equal(t, "mysql://localhost:3306/test", value)
}

// TestApolloSubscriber_LoadValue_KeyNotFound tests error handling for missing keys
func TestApolloSubscriber_LoadValue_KeyNotFound(t *testing.T) {
	client := newMockApolloClient()

	sub := &apolloSubscriber{
		client: client,
		conf: ApolloConf{
			NamespaceName: "application",
			Key:           "non.existent.key",
		},
	}

	err := sub.loadValue()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key not found")
}

// TestApolloSubscriber_HotReload tests configuration hot reload functionality
func TestApolloSubscriber_HotReload(t *testing.T) {
	client := newMockApolloClient()
	client.configCache.Set("feature.enabled", true, 0)

	sub := &apolloSubscriber{
		client: client,
		conf: ApolloConf{
			NamespaceName: "application.json",
			Format:        "json",
		},
	}

	// Initial load
	err := sub.loadValue()
	assert.NoError(t, err)

	// Add listener to track changes
	changeDetected := false
	var mu sync.Mutex
	sub.AddListener(func() {
		mu.Lock()
		changeDetected = true
		mu.Unlock()
	})

	// Simulate configuration change
	client.configCache.Set("feature.enabled", false, 0)
	client.configCache.Set("feature.new", "added", 0)

	// Trigger change event
	sub.handleConfigChange()

	// Wait a bit for async processing
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	assert.True(t, changeDetected, "Listener should be called on config change")
	mu.Unlock()

	// Verify updated value
	value, err := sub.Value()
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(value), &result)
	assert.NoError(t, err)
	assert.Equal(t, false, result["feature.enabled"])
	assert.Equal(t, "added", result["feature.new"])
}

// TestApolloSubscriber_MultipleListeners tests multiple listener registration
func TestApolloSubscriber_MultipleListeners(t *testing.T) {
	client := newMockApolloClient()
	client.configCache.Set("key", "value1", 0)

	sub := &apolloSubscriber{
		client: client,
		conf: ApolloConf{
			NamespaceName: "application.json",
			Format:        "json",
		},
	}

	err := sub.loadValue()
	assert.NoError(t, err)

	// Add multiple listeners
	callCount := 0
	var mu sync.Mutex

	for i := 0; i < 3; i++ {
		sub.AddListener(func() {
			mu.Lock()
			callCount++
			mu.Unlock()
		})
	}

	// Trigger change
	client.configCache.Set("key", "value2", 0)
	sub.handleConfigChange()

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, 3, callCount, "All listeners should be called")
	mu.Unlock()
}

// TestApolloSubscriber_PropertiesFormat tests properties format support
func TestApolloSubscriber_PropertiesFormat(t *testing.T) {
	client := newMockApolloClient()
	client.configCache.Set("server.port", "8080", 0)
	client.configCache.Set("server.host", "localhost", 0)
	client.configCache.Set("database.url", "mysql://localhost:3306", 0)

	sub := &apolloSubscriber{
		client: client,
		conf: ApolloConf{
			NamespaceName: "application.properties",
			Format:        "properties",
		},
	}

	err := sub.loadValue()
	assert.NoError(t, err)

	value, err := sub.Value()
	assert.NoError(t, err)

	// Properties format should be key=value
	assert.Contains(t, value, "server.port=8080")
	assert.Contains(t, value, "server.host=localhost")
	assert.Contains(t, value, "database.url=mysql://localhost:3306")
}

// TestApolloSubscriber_ConcurrentAccess tests thread-safety
func TestApolloSubscriber_ConcurrentAccess(t *testing.T) {
	client := newMockApolloClient()
	client.configCache.Set("counter", 0, 0)

	sub := &apolloSubscriber{
		client: client,
		conf: ApolloConf{
			NamespaceName: "application.json",
			Format:        "json",
		},
	}

	err := sub.loadValue()
	assert.NoError(t, err)

	// Concurrent reads
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, err := sub.Value()
				assert.NoError(t, err)
			}
		}()
	}

	// Concurrent writes (config changes)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			client.configCache.Set("counter", n, 0)
			sub.handleConfigChange()
		}(i)
	}

	wg.Wait()
	// If we get here without deadlock or panic, the test passes
}
