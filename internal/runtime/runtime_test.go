package runtime

import (
	"sync"
	"testing"
)

func resetExtensions() {
	extensionRegistry = make(map[string]map[string]*Function)
}

func TestNormalizeTypeNameCanonicalizesSynonyms(t *testing.T) {
	cases := map[string]string{
		"Int":     "Number",
		"Integer": "Number",
		"Float":   "Number",
		"Bool":    "Boolean",
		"Boolean": "Boolean",
		"String":  "String",
	}
	for input, want := range cases {
		if got := normalizeTypeName(input); got != want {
			t.Fatalf("normalizeTypeName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestExtensionRegistryUsesCanonicalName(t *testing.T) {
	resetExtensions()
	fn := &Function{Name: "size"}
	registerExtension("Int", "size", fn)
	if got, ok := lookupExtension("Number", "size"); !ok || got != fn {
		t.Fatalf("lookupExtension failed after registering canonical name")
	}
	if _, ok := lookupExtension("Number", "missing"); ok {
		t.Fatalf("lookupExtension returned result for unknown method")
	}
}

func TestNewModuleCopiesExports(t *testing.T) {
	exports := map[string]Value{
		"pi":   NewNumber(3.14),
		"name": NewString("selene"),
	}
	module := NewModule("math", exports)
	exports["pi"] = NewNumber(0)
	if module.Exports["pi"].Inspect() != "3.14" {
		t.Fatalf("module exports mutated when source map changed")
	}
	module.Exports["name"] = NewString("other")
	if exports["name"].Inspect() != "selene" {
		t.Fatalf("original map mutated when module exports updated")
	}
}

func TestTaskJoinDeliversResultOnce(t *testing.T) {
	task := NewTask()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		task.deliver(NewString("ok"), nil)
	}()
	v1, err1 := task.Join()
	if err1 != nil {
		t.Fatalf("first Join returned error: %v", err1)
	}
	v2, err2 := task.Join()
	if err2 != nil {
		t.Fatalf("second Join returned error: %v", err2)
	}
	if v1 != v2 {
		t.Fatalf("Join did not return cached result on subsequent calls")
	}
	wg.Wait()
}
