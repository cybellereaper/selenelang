package runtime

import (
	"cmp"
	"maps"
	"slices"
	"strings"
	"sync"
)

const builderMaxReuse = 1 << 12

var builderPool = sync.Pool{New: func() any {
	return &strings.Builder{}
}}

func borrowBuilder() *strings.Builder {
	if b, ok := builderPool.Get().(*strings.Builder); ok && b != nil {
		b.Reset()
		return b
	}
	return &strings.Builder{}
}

func finishBuilder(b *strings.Builder) string {
	result := b.String()
	if b.Cap() > builderMaxReuse {
		*b = strings.Builder{}
	} else {
		b.Reset()
	}
	builderPool.Put(b)
	return result
}

func sortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	if len(m) == 0 {
		return nil
	}
	keys := slices.Collect(maps.Keys(m))
	slices.Sort(keys)
	return keys
}

func inspectFields(name string, fields map[string]Value) string {
	if len(fields) == 0 {
		if name == "" {
			return "{}"
		}
		return name + "{}"
	}

	keys := sortedKeys(fields)

	b := borrowBuilder()
	if name != "" {
		b.WriteString(name)
	}
	b.WriteByte('{')
	b.Grow(len(name) + len(keys)*4)
	for i, key := range keys {
		if i > 0 {
			b.WriteString(", ")
		}
		value := fields[key]
		b.WriteString(key)
		b.WriteString(": ")
		b.WriteString(value.Inspect())
	}
	b.WriteByte('}')
	return finishBuilder(b)
}
