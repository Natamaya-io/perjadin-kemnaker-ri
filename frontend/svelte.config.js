import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		// Setting biar output folder namanya 'build'
		adapter: adapter({
			out: 'build',
			precompress: false
		}),
		output: {
			preloadStrategy: 'modulepreload'
		}
	}
};

export default config;
