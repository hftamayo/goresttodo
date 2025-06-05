package keys

import (
	"sort"
	"strings"
)

// Generator provides a standard way to create cache keys
type Generator struct {
    Prefix string
}

// NewGenerator creates a cache key generator with optional prefix
func NewGenerator(prefix string) *Generator {
    return &Generator{
        Prefix: prefix,
    }
}

// Build creates a cache key with parts joined by underscore
func (g *Generator) Build(parts ...string) string {
    if g.Prefix != "" {
        return g.Prefix + "_" + strings.Join(parts, "_")
    }
    return strings.Join(parts, "_")
}

// ForList builds a key for list operations
func (g *Generator) ForList(params map[string]string) string {
    // Sort the param keys for consistent ordering
    var keys []string
    for k := range params {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    
    var parts []string
    if g.Prefix != "" {
        parts = append(parts, g.Prefix)
    }
    parts = append(parts, "list")
    
    for _, k := range keys {
        parts = append(parts, k+"_"+params[k])
    }
    
    return strings.Join(parts, "_")
}