IMAGE_BACKEND = exercise-backend:latest
IMAGE_FRONTEND = exercise-frontend:latest

.PHONY: build build-backend build-frontend run run-backend run-frontend stop clean

build: build-backend build-frontend

build-backend:
	docker build -t $(IMAGE_BACKEND) ./backend

build-frontend:
	docker build  -t $(IMAGE_FRONTEND) ./frontend

run: run-backend run-frontend

run-backend:
	docker run -d --rm --name exercise-backend \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=exercise \
		-p 8081:8081 \
		-p 5432:5432 $(IMAGE_BACKEND)

run-frontend:
	docker run -d --rm --name exercise-frontend -p 8080:80 $(IMAGE_FRONTEND)

stop:
	docker stop exercise-backend exercise-frontend || true

clean: stop
	docker rmi -f $(IMAGE_BACKEND) $(IMAGE_FRONTEND) || true
