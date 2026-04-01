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

			// Ambil body sebagai Buffer (Node.js native) agar lebih stabil saat dikirim balik
			const responseData = await response.arrayBuffer();
			const buffer = Buffer.from(responseData);
			
			console.log(`[Proxy Debug] ${url} -> ${buffer.length} bytes`);

			const proxyHeaders = new Headers();
			let contentType = response.headers.get('content-type');
			
			// --- IDM CLOAKING LOGIC ---
			// Jika tipe adalah PDF, kita samarkan jadi octet-stream 
			// supaya tidak dicuri oleh Internet Download Manager (IDM)
			if (contentType && contentType.includes('application/pdf')) {
				proxyHeaders.set('content-type', 'application/octet-stream');
			} else if (contentType) {
				proxyHeaders.set('content-type', contentType);
			}
			
			const contentDisposition = response.headers.get('content-disposition');
			if (contentDisposition) proxyHeaders.set('content-disposition', contentDisposition);

			// Paksa no-cache agar browser tidak menyimpan respons error 0-byte sebelumnya
			proxyHeaders.set('cache-control', 'no-cache, no-store, must-revalidate');
			proxyHeaders.set('pragma', 'no-cache');
			proxyHeaders.set('expires', '0');

			return new Response(buffer, {
				status: response.status,
				statusText: response.statusText,
				headers: proxyHeaders
			});
		} catch (err: any) {
			// Log ini akan muncul di 'podman logs <frontend_container>'
			console.error(`[Proxy Error] ${event.request.method} ${url} ->`, err.message);
			return new Response(`Proxy Error: ${err.message}`, { status: 502 });
		}
	}

	return resolve(event);
};
