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

package main

import (
	"context"
	"log"
	"os"

	"google.golang.org/genai"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/artifact"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/full"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/geminitool"
)

func main() {
	ctx := context.Background()

	// 1. Handle Help
	if len(os.Args) > 1 && os.Args[1] == "help" {
		l := full.NewLauncher()
		_ = l.Execute(ctx, nil, os.Args[1:])
		return
	}

	// 2. Initialize Model (Gemini 2.0 Flash)
	apiKey := os.Getenv("GOOGLE_API_KEY")
	var llm model.LLM
	var err error

	if apiKey == "" || apiKey == "mock" {
		log.Println("Starting in MOCK mode. Set GOOGLE_API_KEY for real Gemini 2.0 interaction.")
		llm = &MockModel{}
	} else {
		llm, err = gemini.NewModel(ctx, "gemini-2.0-flash", &genai.ClientConfig{
			APIKey: apiKey,
		})
		if err != nil {
			log.Fatalf("Failed to create model: %v", err)
		}
	}

	// 3. Define the Primary Agent
	primaryAgent, err := llmagent.New(llmagent.Config{
		Name:        "starter_agent",
		Model:       llm,
		Description: "A versatile starter agent built with ADK 2.0 principles.",
		Instruction: "You are a helpful and precise assistant. Use the available tools to provide accurate information.",
		Tools: []tool.Tool{
			geminitool.GoogleSearch{},
		},
	})
	if err != nil {
		log.Fatalf("Failed to create primary agent: %v", err)
	}

	// 4. Setup Services
	sessionService := session.InMemoryService()
	artifactService := artifact.InMemoryService()

	// 5. Configure Launcher
	config := &launcher.Config{
		AgentLoader:     agent.NewSingleLoader(primaryAgent),
		SessionService:  sessionService,
		ArtifactService: artifactService,
	}

	// 6. Launch
	l := full.NewLauncher()
	if err = l.Execute(ctx, config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}
