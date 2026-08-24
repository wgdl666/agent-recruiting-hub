.PHONY: dev backend frontend build seed run install deploy

ROOT := $(shell pwd)

install:
	cd backend && go mod tidy
	cd frontend && npm install

backend:
	cd backend && go run . -root $(ROOT)

frontend:
	cd frontend && npm run dev

build:
	cd frontend && npm run build
	cd backend && go build -o ../bin/hub .

run: build
	./bin/hub -root $(ROOT)

seed:
	chmod +x scripts/hub
	curl -s -X POST http://localhost:8080/api/import/seed | python3 -m json.tool

rescreen:
	curl -s -X POST http://localhost:8080/api/rescreen | python3 -m json.tool

feishu:
	curl -s -X POST http://localhost:8080/api/export/feishu | python3 -m json.tool

deploy:
	chmod +x scripts/deploy-mirror.sh
	./scripts/deploy-mirror.sh

# 默认部署到 mirror_zby → https://recruit.wgdl.tech

dev:
	@echo "Run in two terminals: make backend && make frontend"
	@echo "Or production: make run -> http://localhost:8080"
