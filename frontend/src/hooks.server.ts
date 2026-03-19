import type { Handle } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';

export const handle: Handle = async ({ event, resolve }) => {
	// Proxy request ke backend untuk path /api dan /uploads
	if (event.url.pathname.startsWith('/api') || event.url.pathname.startsWith('/uploads')) {
		const target = env.INTERNAL_API_URL || 'http://backend:8081';
		const url = `${target}${event.url.pathname}${event.url.search}`;

		// Persiapkan headers, hapus header yang bisa menyebabkan masalah proxy
		const headers = new Headers(event.request.headers);
		headers.delete('host');
		headers.delete('connection');

		try {
			// Gunakan clone() agar request asli tetap tersedia jika dibutuhkan SvelteKit
			const requestClone = event.request.clone();
			
			const fetchOptions: RequestInit = {
				method: event.request.method,
				headers: headers,
				// duplex: 'half' wajib ada untuk mem-proxy body stream di Node.js
				// @ts-ignore
				duplex: 'half'
			};

			if (event.request.method !== 'GET' && event.request.method !== 'HEAD') {
				fetchOptions.body = requestClone.body;
			}

			const response = await fetch(url, fetchOptions);

			// Ambil body sebagai ArrayBuffer untuk memastikan integritas data (terutama file upload)
			const responseData = await response.arrayBuffer();

			return new Response(responseData, {
				status: response.status,
				statusText: response.statusText,
				headers: response.headers
			});
		} catch (err: any) {
			// Log ini akan muncul di 'podman logs <frontend_container>'
			console.error(`[Proxy Error] ${event.request.method} ${url} ->`, err.message);
			return new Response(`Proxy Error: ${err.message}`, { status: 502 });
		}
	}

	return resolve(event);
};
