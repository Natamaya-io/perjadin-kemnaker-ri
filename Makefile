.PHONY: dev down logs build stop

dev:
	doppler run -- podman compose up -d --build

stop:
	podman compose stop

down:
	podman compose down

logs:
	podman compose logs -f
