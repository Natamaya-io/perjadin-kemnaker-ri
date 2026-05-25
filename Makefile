.PHONY: dev down logs build stop watch

dev:
	doppler run -- podman compose up -d --build

build:
	podman compose build backend
	podman compose build frontend

stop:
	podman compose stop

down:
	podman compose down

logs:
	podman compose logs -f

watch:
	@echo "Starting services initially..."
	$(MAKE) down
	$(MAKE) dev
	@echo "Watching for changes to trigger restart..."
	@while true; do \
		inotifywait -q -r -e modify,create,delete,move --exclude '(\.git|node_modules|__pycache__|\.svelte-kit)' . ; \
		echo "Changes detected, restarting services..."; \
		$(MAKE) down; \
		$(MAKE) dev; \
	done
