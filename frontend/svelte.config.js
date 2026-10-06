import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		csrf: {
			checkOrigin: true,
			trustedOrigins: ['http://localhost:3000', 'http://127.0.0.1:3000', 'http://100.115.101.14:3000', 'https://perjadin.kemnaker.go.id', 'http://perjadin.kemnaker.go.id', 'https://sinurdin.gatsu51.com']
		},
		// Setting SPA Static
		adapter: adapter({ 
			pages: 'build',
			assets: 'build',
			fallback: 'index.html',
			precompress: false,
			strict: true 
		}),
		output: {
			preloadStrategy: 'preload-mjs'
		}
	}
};

export default config;
