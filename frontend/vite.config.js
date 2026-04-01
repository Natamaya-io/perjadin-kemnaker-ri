import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		host: true,
		proxy: {
			'/api': {
				target: process.env.BACKEND_URL || 'http://127.0.0.1:8081',
				changeOrigin: true
			},
			'/uploads': {
				target: process.env.BACKEND_URL || 'http://127.0.0.1:8081',
				changeOrigin: true
			}
		}
	}
});
