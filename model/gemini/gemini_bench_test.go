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

package gemini

import (
	"net/http"
	"testing"
)

type dummyTransport struct{}

func (d *dummyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: 200}, nil
}

func BenchmarkMergeHeadersInterceptor_SingleValue(b *testing.B) {
	interceptor := &mergeHeadersInterceptor{
		base: &dummyTransport{},
	}
	req, _ := http.NewRequest("POST", "https://example.com", nil)
	req.Header = http.Header{
		"X-Goog-Api-Client": []string{"google-adk/v2.0.0 gl-go/1.25.0"},
		"User-Agent":        []string{"google-adk/v2.0.0 gl-go/1.25.0"},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		_, _ = interceptor.RoundTrip(req)
	}
}

func BenchmarkMergeHeadersInterceptor_MultipleValues(b *testing.B) {
	interceptor := &mergeHeadersInterceptor{
		base: &dummyTransport{},
	}
	req, _ := http.NewRequest("POST", "https://example.com", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		b.StopTimer()
		req.Header = http.Header{
			"X-Goog-Api-Client": []string{"google-adk/v2.0.0 gl-go/1.25.0", "genai-go/1.63.0"},
			"User-Agent":        []string{"google-adk/v2.0.0 gl-go/1.25.0", "genai-go/1.63.0"},
		}
		b.StartTimer()

		_, _ = interceptor.RoundTrip(req)
	}
}
