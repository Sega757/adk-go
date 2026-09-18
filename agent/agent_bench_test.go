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

package agent

import (
	"context"
	"testing"

	"google.golang.org/adk/v2/session"
)

func BenchmarkPrepareEventActions_Nil(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		actions := prepareEventActions(nil)
		if actions == nil {
			b.Fatal("unexpected nil actions")
		}
	}
}

func BenchmarkPrepareEventActions_Existing(b *testing.B) {
	b.ReportAllocs()
	actions := &session.EventActions{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		act := prepareEventActions(actions)
		if act == nil {
			b.Fatal("unexpected nil actions")
		}
	}
}

func BenchmarkRunBeforeAgentCallbacks_NoCallbacks(b *testing.B) {
	b.ReportAllocs()
	svc := session.InMemoryService()
	resp, err := svc.Create(context.Background(), &session.CreateRequest{AppName: "app", UserID: "user"})
	if err != nil {
		b.Fatal(err)
	}
	ag, err := New(Config{
		Name: "test_agent",
	})
	if err != nil {
		b.Fatal(err)
	}
	ic := &invocationContext{
		Context: context.Background(),
		agent:   ag,
		session: resp.Session,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ev, err := runBeforeAgentCallbacks(ic)
		if err != nil || ev != nil {
			b.Fatal("unexpected result")
		}
	}
}

func BenchmarkRunAfterAgentCallbacks_NoCallbacks(b *testing.B) {
	b.ReportAllocs()
	svc := session.InMemoryService()
	resp, err := svc.Create(context.Background(), &session.CreateRequest{AppName: "app", UserID: "user"})
	if err != nil {
		b.Fatal(err)
	}
	ag, err := New(Config{
		Name: "test_agent",
	})
	if err != nil {
		b.Fatal(err)
	}
	ic := &invocationContext{
		Context: context.Background(),
		agent:   ag,
		session: resp.Session,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ev, err := runAfterAgentCallbacks(ic)
		if err != nil || ev != nil {
			b.Fatal("unexpected result")
		}
	}
}
