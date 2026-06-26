# ADK 2.0 Starter Kit

This starter kit provides a robust, production-ready foundation for building AI agents using the Google Agent Development Kit (ADK) for Go.

## 🚀 Quick Start

### 1. Run in Console Mode (Mock)
Interactively chat with your agent in the terminal without needing an API key.
```bash
GOOGLE_API_KEY=mock go run ./starter-kit console
```

### 2. Run with Web UI (Mock)
Launch a full web server with a built-in UI to test your agents.
```bash
GOOGLE_API_KEY=mock go run ./starter-kit web api webui
```
Then open [http://localhost:8080/ui/](http://localhost:8080/ui/) in your browser.

## 🛠️ Customization

### Adding Tools
To add more capabilities, register them in the `Tools` slice in `main.go`:
```go
Tools: []tool.Tool{
    geminitool.GoogleSearch{},
    // Add your custom tools here
},
```

### Using Real Gemini 2.0
Set your Google AI API Key to enable real model interactions:
```bash
export GOOGLE_API_KEY="your-api-key"
go run ./starter-kit console
```

## 🚢 Deployment

Deploy your agent to Google Cloud Run with one command using the `adkgo` CLI:

```bash
go run ./cmd/adkgo deploy cloudrun   --project_name YOUR_PROJECT_ID   --service_name starter-agent   --entry_point_path ./starter-kit
```

*Note: Ensure you have `gcloud` installed and authenticated.*
