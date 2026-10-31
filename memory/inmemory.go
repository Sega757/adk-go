// Copyright 2025 Google LLC
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

package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"google.golang.org/genai"

	"google.golang.org/adk/v2/session"
)

// InMemoryService returns a new in-memory implementation of the memory service. Thread-safe.
func InMemoryService() Service {
	return &inMemoryService{
		store: make(map[key]map[sessionID][]value),
	}
}

type key struct {
	appName, userID string
}

type sessionID string

type value struct {
	id             string
	content        *genai.Content
	author         string
	timestamp      time.Time
	customMetadata map[string]any

	// precomputed set of words in the content for simple keyword matching.
	words map[string]struct{}
}

// inMemoryService is an in-memory implementation of Service.
type inMemoryService struct {
	mu    sync.RWMutex
	store map[key]map[sessionID][]value
}

func (s *inMemoryService) AddSessionToMemory(ctx context.Context, curSession session.Session) error {
	var values []value

	events := curSession.Events()
	numEvents := events.Len()
	for i := 0; i < numEvents; i++ {
		event := events.At(i)
		if event.LLMResponse.Content == nil {
			continue
		}

		// Lazily allocate the words map only when non-empty text parts are encountered,
		// and populate directly to avoid intermediate map allocations and maps.Copy churn.
		var words map[string]struct{}
		for _, part := range event.LLMResponse.Content.Parts {
			if part.Text == "" {
				continue
			}

			if words == nil {
				words = make(map[string]struct{})
			}
			addWords(words, part.Text)
		}

		if len(words) == 0 {
			continue
		}

		values = append(values, value{
			id:             event.ID,
			content:        event.LLMResponse.Content,
			author:         event.Author,
			timestamp:      event.Timestamp,
			customMetadata: event.CustomMetadata,
			words:          words,
		})
	}

	k := key{
		appName: curSession.AppName(),
		userID:  curSession.UserID(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.store[k]
	if !ok {
		v = map[sessionID][]value{}
		s.store[k] = v
	}

	sid := sessionID(curSession.ID())
	v[sid] = values
	return nil
}

func (s *inMemoryService) SearchMemory(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	k := key{
		appName: req.AppName,
		userID:  req.UserID,
	}

	s.mu.RLock()
	values, ok := s.store[k]
	s.mu.RUnlock()
	if !ok || len(values) == 0 {
		return &SearchResponse{}, nil
	}

	queryWords := extractWords(req.Query)
	if len(queryWords) == 0 {
		return &SearchResponse{}, nil
	}

	res := &SearchResponse{}

	for _, events := range values {
		for _, e := range events {
			if checkMapsIntersect(e.words, queryWords) {
				res.Memories = append(res.Memories, Entry{
					ID:             e.id,
					Content:        e.content,
					Author:         e.author,
					Timestamp:      e.timestamp,
					CustomMetadata: e.customMetadata,
				})
			}
		}
	}

	return res, nil
}

func checkMapsIntersect(m1, m2 map[string]struct{}) bool {
	if len(m1) == 0 || len(m2) == 0 {
		return false
	}

	// Iterate over the smaller map.
	if len(m1) > len(m2) {
		m1, m2 = m2, m1
	}

	for k := range m1 {
		if _, ok := m2[k]; ok {
			return true
		}
	}

	return false
}

// addWords parses space-delimited words from text and records them into words in lowercase.
// Using strings.IndexByte avoids iterator closure and slice allocations.
func addWords(words map[string]struct{}, text string) {
	for text != "" {
		idx := strings.IndexByte(text, ' ')
		var s string
		if idx < 0 {
			s = text
			text = ""
		} else {
			s = text[:idx]
			text = text[idx+1:]
		}
		if s == "" {
			continue
		}
		words[strings.ToLower(s)] = struct{}{}
	}
}

func extractWords(text string) map[string]struct{} {
	res := make(map[string]struct{})
	addWords(res, text)
	return res
}
