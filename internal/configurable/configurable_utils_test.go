// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package configurable

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"google.golang.org/adk/v2/tool"
)

func resetRegistries(t *testing.T) {
	t.Helper()
	registryMu.Lock()
	oldCallbacks := make(map[string]any, len(callbackRegistry))
	for k, v := range callbackRegistry {
		oldCallbacks[k] = v
	}
	oldTools := make(map[string]any, len(toolRegistry))
	for k, v := range toolRegistry {
		oldTools[k] = v
	}
	registryMu.Unlock()

	t.Cleanup(func() {
		registryMu.Lock()
		defer registryMu.Unlock()
		callbackRegistry = oldCallbacks
		toolRegistry = oldTools
	})
}

func TestRegisterCallback(t *testing.T) {
	resetRegistries(t)

	t.Run("EmptyNameValidation", func(t *testing.T) {
		err := RegisterCallback("", func() {})
		if err == nil {
			t.Errorf("expected error when registering callback with empty name, got nil")
		}
	})

	t.Run("NilCallbackValidation", func(t *testing.T) {
		err := RegisterCallback("test_nil_cb", nil)
		if err == nil {
			t.Errorf("expected error when registering nil callback, got nil")
		}
		_, resolveErr := ResolveCallbackReference(context.Background(), "test_nil_cb")
		if resolveErr == nil {
			t.Errorf("expected nil callback not to be found in registry")
		}
	})

	t.Run("TypeCompatibility", func(t *testing.T) {
		type dummyStruct struct {
			Name string
		}

		dummyFn := func(ctx context.Context, input string) (string, error) {
			return "hello " + input, nil
		}
		dummyObj := dummyStruct{Name: "test"}
		dummyClosure := func() int { return 42 }

		tests := []struct {
			name string
			val  any
		}{
			{name: "test_func", val: dummyFn},
			{name: "test_struct", val: dummyObj},
			{name: "test_closure", val: dummyClosure},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				err := RegisterCallback(tc.name, tc.val)
				if err != nil {
					t.Fatalf("unexpected error registering callback %s: %v", tc.name, err)
				}
				resolved, err := ResolveCallbackReference(context.Background(), tc.name)
				if err != nil {
					t.Fatalf("unexpected error resolving callback %s: %v", tc.name, err)
				}
				if resolved == nil {
					t.Fatalf("expected resolved callback for %s to be non-nil", tc.name)
				}
			})
		}
	})

	t.Run("DuplicateRegistrationRejection", func(t *testing.T) {
		name := "test_dup_cb"
		firstCb := func() string { return "first" }
		secondCb := func() string { return "second" }

		err := RegisterCallback(name, firstCb)
		if err != nil {
			t.Fatalf("failed initial registration: %v", err)
		}

		err = RegisterCallback(name, secondCb)
		if err == nil {
			t.Fatalf("expected error on duplicate callback registration, got nil")
		}

		resolved, err := ResolveCallbackReference(context.Background(), name)
		if err != nil {
			t.Fatalf("failed to resolve callback after duplicate attempt: %v", err)
		}
		fn, ok := resolved.(func() string)
		if !ok || fn() != "first" {
			t.Fatalf("expected original callback to remain registered")
		}
	})
}

func TestRegisterToolsetFactory(t *testing.T) {
	resetRegistries(t)

	t.Run("EmptyNameValidation", func(t *testing.T) {
		factory := func(ctx context.Context, args map[string]any) (tool.Toolset, error) {
			return nil, nil
		}
		err := RegisterToolsetFactory("", factory)
		if err == nil {
			t.Errorf("expected error when registering toolset factory with empty name, got nil")
		}
	})

	t.Run("NilFactoryValidation", func(t *testing.T) {
		err := RegisterToolsetFactory("test_nil_toolset", nil)
		if err == nil {
			t.Errorf("expected error when registering nil toolset factory, got nil")
		}
	})

	t.Run("HappyPathAndDuplicateRejection", func(t *testing.T) {
		name := "test_toolset_factory"
		factory1 := func(ctx context.Context, args map[string]any) (tool.Toolset, error) {
			return nil, nil
		}
		factory2 := func(ctx context.Context, args map[string]any) (tool.Toolset, error) {
			return nil, fmt.Errorf("factory2")
		}

		err := RegisterToolsetFactory(name, factory1)
		if err != nil {
			t.Fatalf("unexpected error registering toolset factory: %v", err)
		}

		err = RegisterToolsetFactory(name, factory2)
		if err == nil {
			t.Fatalf("expected error on duplicate toolset factory registration, got nil")
		}

		_, toolset, err := ResolveToolReference(context.Background(), name, nil)
		if err != nil {
			t.Fatalf("unexpected error resolving registered toolset factory: %v", err)
		}
		if toolset != nil {
			t.Fatalf("expected toolset to be nil from dummy factory")
		}
	})

	t.Run("CrossTypeCollisionWithToolFactory", func(t *testing.T) {
		toolName := "cross_collision_tool"
		toolFactory := func(ctx context.Context, args map[string]any) (tool.Tool, error) {
			return nil, nil
		}
		toolsetFactory := func(ctx context.Context, args map[string]any) (tool.Toolset, error) {
			return nil, nil
		}

		if err := RegisterToolFactory(toolName, toolFactory); err != nil {
			t.Fatalf("unexpected error registering tool factory: %v", err)
		}

		err := RegisterToolsetFactory(toolName, toolsetFactory)
		if err == nil {
			t.Fatalf("expected error registering toolset factory with duplicate tool name, got nil")
		}

		toolsetName := "cross_collision_toolset"
		if err := RegisterToolsetFactory(toolsetName, toolsetFactory); err != nil {
			t.Fatalf("unexpected error registering toolset factory: %v", err)
		}

		err = RegisterToolFactory(toolsetName, toolFactory)
		if err == nil {
			t.Fatalf("expected error registering tool factory with duplicate toolset name, got nil")
		}
	})
}

func TestConcurrentRegistration(t *testing.T) {
	resetRegistries(t)

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	for i := 0; i < numGoroutines; i++ {
		keyName := fmt.Sprintf("concurrent_cb_%d", i)
		go func(k string) {
			defer wg.Done()
			_ = RegisterCallback(k, func() {})
		}(keyName)

		collidingKey := "concurrent_colliding_cb"
		go func() {
			defer wg.Done()
			_ = RegisterCallback(collidingKey, func() {})
		}()
	}

	wg.Wait()
}
