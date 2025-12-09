// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package entrest

import (
	"cmp"
	"encoding/json"
	"fmt"
	"slices"
	"sync"

	"entgo.io/ent/entc/gen"
	"github.com/fatih/structtag"
)

// EdgeHasOperation checks if an edge has an operation, with inheritance from the parent type.
// The lookup order is:
//  1. Edge's explicit operations (if set via WithIncludeOperations on the edge)
//  2. Parent type's explicit operations (if set via WithIncludeOperations on the schema)
//  3. Global default operations from config
//
// This ensures that when you enable an operation on an entity (e.g., Upsert), its edges
// automatically participate in that operation without requiring explicit annotations on each edge.
func EdgeHasOperation(edge *gen.Edge, parentType *gen.Type, config *Config, op Operation) bool {
	edgeAnt := GetAnnotation(edge)

	// If edge has explicit operations, use those
	if edgeAnt.Operations != nil {
		// Check exclusions
		if len(edgeAnt.ExcludedOperations) > 0 && slices.Contains(edgeAnt.ExcludedOperations, op) {
			return false
		}
		return slices.Contains(edgeAnt.Operations, op)
	}

	// If edge has explicit exclusions but no explicit inclusions, check parent then defaults
	if len(edgeAnt.ExcludedOperations) > 0 && slices.Contains(edgeAnt.ExcludedOperations, op) {
		return false
	}

	// Check parent type's operations
	parentAnt := GetAnnotation(parentType)
	if parentAnt.Operations != nil {
		return slices.Contains(parentAnt.Operations, op)
	}

	// Finally, fall back to global defaults
	return slices.Contains(config.DefaultOperations, op)
}

// ptr returns a pointer to the given value. Should only be used for primitives.
func ptr[T any](v T) *T {
	return &v
}

// memoize memoizes the provided function, so that it is only called once for each
// input.
func memoize[K comparable, V any](fn func(K) V) func(K) V {
	var mu sync.RWMutex
	cache := map[K]V{}

	return func(in K) V {
		mu.RLock()
		if cached, ok := cache[in]; ok {
			mu.RUnlock()
			return cached
		}
		mu.RUnlock()

		mu.Lock()
		defer mu.Unlock()

		cache[in] = fn(in)
		return cache[in]
	}
}

// sliceToRawMessage returns a slice of json.RawMessage from a slice of T. Panics
// if any of the values cannot be marshaled to JSON.
func sliceToRawMessage[T any](v []T) []json.RawMessage {
	r := make([]json.RawMessage, len(v))
	var err error
	for i, v := range v {
		r[i], err = json.Marshal(v)
		if err != nil {
			panic(fmt.Sprintf("failed to marshal %v: %v", v, err))
		}
	}
	return r
}

// appendCompactFunc returns a copy of orig with newv appended to it, but only if newv does
// not already exist in orig. fn is used to determine if two values are equal.
func appendCompactFunc[T any](orig, newv []T, fn func(oldv, newv T) (matches bool)) []T {
	for _, v := range newv {
		var found bool
		for _, ov := range orig {
			if fn(ov, v) {
				found = true
				break
			}
		}
		if !found {
			orig = append(orig, v)
		}
	}
	return orig
}

// appendCompact returns a copy of orig with newv appended to it, but only if newv does
// not already exist in orig. T must be comparable.
func appendCompact[T comparable](orig, newv []T) []T {
	return appendCompactFunc(orig, newv, func(oldv, newv T) bool {
		return oldv == newv
	})
}

// sliceCompact is similasr to slices.Compact, but it keeps the original slice ordering.
func sliceCompact[T comparable](orig []T) (compacted []T) {
	// Start with an empty slice.
	return appendCompact(compacted, orig)
}

// mergeMap returns a copy of orig with newv merged into it, but only if
// newv does not already exist in orig. If orig is nil, this will panic, as we cannot
// merge into a nil map without returning a new map.
func mergeMap[K comparable, V any](overlap bool, orig, newv map[K]V) error {
	if orig == nil {
		panic("orig is nil")
	}
	if newv == nil {
		return nil
	}

	for k, v := range newv {
		_, ok := orig[k]
		if !overlap && ok {
			return fmt.Errorf("key %v already exists in original map", k)
		}

		if !ok || overlap {
			orig[k] = v
			continue
		}
	}
	return nil
}

// sliceOr returns the provided default value(s) if the given value is nil. This is like
// [cmp.Or] for slices.
func sliceOr[T any](v []T, defaults ...[]T) []T {
	if len(v) == 0 {
		for i := range defaults {
			if len(defaults[i]) > 0 {
				return defaults[i]
			}
		}
	}
	return v
}

// mapKeys returns the keys of the map m, sorted.
func mapKeys[M ~map[K]V, K cmp.Ordered, V any](m M) []K {
	r := make([]K, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	slices.Sort(r)
	return r
}

// ToEnum returns a slice of json.RawMessage from a slice of T. This is useful when
// using the [WithSchema] annotation.
func ToEnum[T any](values []T) ([]json.RawMessage, error) {
	results := make([]json.RawMessage, len(values))
	var err error
	for i, e := range values {
		results[i], err = json.Marshal(e)
		if err != nil {
			return nil, err
		}
	}
	return results, nil
}

// intersect returns the intersection between two slices.
func intersect[T comparable, S ~[]T](a, b S) S {
	result := S{}
	seen := map[T]struct{}{}
	for i := range a {
		seen[a[i]] = struct{}{}
	}
	for i := range b {
		if _, ok := seen[b[i]]; ok {
			result = append(result, b[i])
		}
	}
	return result
}

// intersectSorted returns the intersection between two slices, and sorts the result.
func intersectSorted[T cmp.Ordered, S ~[]T](a, b S) S {
	out := intersect(a, b)
	slices.Sort(out)
	return out
}

func patchJSONTag(g *gen.Graph) error {
	for _, node := range g.Nodes {
		for _, field := range node.Fields {
			if field.StructTag == `json:"-"` {
				continue
			}
			tags, err := structtag.Parse(field.StructTag)
			if err != nil {
				return fmt.Errorf("failed to parse struct tag for field %q: %w", field.Name, err)
			}
			tags.DeleteOptions("json", "omitempty")
			field.StructTag = tags.String()
		}
	}
	return nil
}
