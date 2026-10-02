package telemetry

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type EventBusPayload struct {
	EventID   string                 `json:"event_id"`
	Timestamp string                 `json:"timestamp"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Payload   map[string]interface{} `json:"payload"`
}

type Producer struct {
	queue    chan EventBusPayload
	filepath string
	file     *os.File
	mu       sync.Mutex
}

func InitProducer(path string) (*Producer, error) {
	f, err := openAppend(path)
	if err != nil {
		return nil, err
	}

	p := &Producer{
		queue:    make(chan EventBusPayload, 10000), // Буфер для защиты R-E-K от I/O лагов
		filepath: path,
		file:     f,
	}

	go p.worker()
	go p.listenForLogRotate() // Слушаем SIGHUP от rotate.sh

	return p, nil
}

func (p *Producer) Emit(event EventBusPayload) {
	event.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	select {
	case p.queue <- event:
	default:
		log.Printf("[FLAT.SYNTH] WARNING: Telemetry queue full, dropping event %s", event.EventID)
	}
}

func (p *Producer) worker() {
	for event := range p.queue {
		data, err := json.Marshal(event)
		if err != nil {
			continue
		}

		p.mu.Lock()
		p.file.Write(data)
		p.file.Write([]byte("\n"))
		p.mu.Unlock()
	}
}

func (p *Producer) listenForLogRotate() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)

	for range sigChan {
		p.mu.Lock()
		p.file.Close()
		f, err := openAppend(p.filepath)
		if err == nil {
			p.file = f
			log.Println("[FLAT.SYNTH] Telemetry file rotated successfully")
		} else {
			log.Printf("[FLAT.SYNTH] ERROR reopening telemetry file: %v", err)
		}
		p.mu.Unlock()
	}
}

func openAppend(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}
