.PHONY: dev down logs build stop watch

dev:
	@echo "Starting development environment (SOTA: Bind Mounts & Native HMR)..."
	doppler run -- podman compose up -d

build:
	podman compose build backend
	podman compose build frontend

stop:
	podman compose stop

down:
	doppler run -- podman compose down -v

logs:
	podman compose logs -f

watch:
	@echo "⚠️ TARGET 'watch' KINI USANG (DEPRECATED) ⚠️"
	@echo "Sistem menggunakan arsitektur SOTA (Vite HMR & Go Air)."
	@echo "Anda tidak perlu melakukan restart manual. Kode akan ter-update otomatis dalam hitungan milidetik saat Anda melakukan Save."
	@echo "Menjalankan target 'dev'..."
	$(MAKE) dev

deploy-staging-fast:
	@echo "🚀 --- SOTA Deploy: Build → GHCR Push → Server Pull ---"
	@echo "📦 [1/4] Building backend image..."
	cd backend && podman build -t ghcr.io/myceldev-com/perjadin-kemnaker-ri-backend:staging .
	@echo "📦 [2/4] Building frontend image (cache-optimized)..."
	cd frontend && podman build -t ghcr.io/myceldev-com/perjadin-kemnaker-ri-frontend:staging .
	@echo "⬆️  [3/4] Pushing ONLY changed layers to GHCR..."
	podman push ghcr.io/myceldev-com/perjadin-kemnaker-ri-backend:staging
	podman push ghcr.io/myceldev-com/perjadin-kemnaker-ri-frontend:staging
	@echo "🔄 [4/4] Server pulling new image and restarting..."
	ssh gatsu51@100.115.101.14 "podman pull ghcr.io/myceldev-com/perjadin-kemnaker-ri-backend:staging && podman pull ghcr.io/myceldev-com/perjadin-kemnaker-ri-frontend:staging && systemctl --user restart perjadin-stg-backend.service perjadin-stg-frontend.service"
	@echo "✅ Deploy selesai!"

deploy-production-fast:
	@echo "🚀 --- PRODUCTION SOTA Deploy: Build → GHCR Push → Server Pull ---"
	@echo "📦 [1/4] Building backend image..."
	cd backend && podman build -t ghcr.io/myceldev-com/perjadin-kemnaker-ri-backend:latest .
	@echo "📦 [2/4] Building frontend image (cache-optimized)..."
	cd frontend && podman build -t ghcr.io/myceldev-com/perjadin-kemnaker-ri-frontend:latest .
	@echo "⬆️  [3/4] Pushing ONLY changed layers to GHCR..."
	podman push ghcr.io/myceldev-com/perjadin-kemnaker-ri-backend:latest
	podman push ghcr.io/myceldev-com/perjadin-kemnaker-ri-frontend:latest
	@echo "🔄 [4/4] Server pulling new image and restarting..."
	ssh gatsu51@100.115.101.14 "podman pull ghcr.io/myceldev-com/perjadin-kemnaker-ri-backend:latest && podman pull ghcr.io/myceldev-com/perjadin-kemnaker-ri-frontend:latest && systemctl --user restart perjadin-backend.service perjadin-frontend.service"
	@echo "✅ Deploy production selesai!"
