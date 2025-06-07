package services

import (
	"fmt"
	"sync"
	"time"
)

type ToolResult struct {
	Data      interface{}
	Timestamp time.Time
	Error     error
}

type ToolRegistry struct {
	mu       sync.RWMutex
	tools    map[string]Tool
	cache    map[string]ToolResult
	cacheTTL time.Duration
}

type Tool interface {
	Name() string
	Description() string
	Execute(params map[string]interface{}) (interface{}, error)
}

func NewToolRegistry(cacheTTL time.Duration) *ToolRegistry {
	return &ToolRegistry{
		tools:    make(map[string]Tool),
		cache:    make(map[string]ToolResult),
		cacheTTL: cacheTTL,
	}
}

func (tr *ToolRegistry) RegisterTool(tool Tool) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.tools[tool.Name()] = tool
}

func (tr *ToolRegistry) GetTool(name string) (Tool, bool) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	tool, exists := tr.tools[name]
	return tool, exists
}

func (tr *ToolRegistry) ListTools() []Tool {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	tools := make([]Tool, 0, len(tr.tools))
	for _, tool := range tr.tools {
		tools = append(tools, tool)
	}
	return tools
}

func (tr *ToolRegistry) ExecuteTool(name string, params map[string]interface{}) (interface{}, error) {
	// Check cache first
	cacheKey := tr.generateCacheKey(name, params)
	if result, exists := tr.getCachedResult(cacheKey); exists {
		return result.Data, result.Error
	}

	// Get and execute tool
	tool, exists := tr.GetTool(name)
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}

	result, err := tool.Execute(params)

	// Cache result
	tr.cacheResult(cacheKey, result, err)

	return result, err
}

func (tr *ToolRegistry) generateCacheKey(name string, params map[string]interface{}) string {
	// Simple cache key generation - can be enhanced based on needs
	return fmt.Sprintf("%s:%v", name, params)
}

func (tr *ToolRegistry) getCachedResult(key string) (ToolResult, bool) {
	tr.mu.RLock()
	defer tr.mu.RUnlock()
	result, exists := tr.cache[key]
	if !exists {
		return ToolResult{}, false
	}
	if time.Since(result.Timestamp) > tr.cacheTTL {
		delete(tr.cache, key)
		return ToolResult{}, false
	}
	return result, true
}

func (tr *ToolRegistry) cacheResult(key string, data interface{}, err error) {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.cache[key] = ToolResult{
		Data:      data,
		Timestamp: time.Now(),
		Error:     err,
	}
}

func (tr *ToolRegistry) ClearCache() {
	tr.mu.Lock()
	defer tr.mu.Unlock()
	tr.cache = make(map[string]ToolResult)
}
