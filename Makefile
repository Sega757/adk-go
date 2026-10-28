.PHONY: test telemetry-setup telemetry-up test-integration

# Существующие тесты ADK + тесты телеметрии
test:
	go test -v ./...

# Подготовка Python-окружения
telemetry-setup:
	cd telemetry && pip install -r requirements.txt

# Запуск консьюмера
telemetry-up: telemetry-setup
	python3 telemetry/transponder.py

# Сквозной тест (отправка события -> проверка)
test-integration:
	@echo "=> Running ADK integration tests..."
	go test -v ./internal/telemetry -run TestProducer_EmitAndWrite
	@echo "=> Integration PASSED."
