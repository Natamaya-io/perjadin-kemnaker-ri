import type { Handle } from '@sveltejs/kit';
import { env } from '$env/dynamic/private';

export const handle: Handle = async ({ event, resolve }) => {
	if (event.url.pathname.startsWith('/api')) {
		const target = env.INTERNAL_API_URL || 'http://backend:8081';
		const url = `${target}${event.url.pathname}${event.url.search}`;

		// Filter headers
		const headers = new Headers(event.request.headers);
		headers.delete('host');
		headers.delete('connection');
		// Remove origin to allow backend to accept request from "server" if strict
		// headers.delete('origin'); 

		try {
			const options: RequestInit = {
				method: event.request.method,
				headers: headers,
				// @ts-ignore
				body: event.request.method !== 'GET' && event.request.method !== 'HEAD' ? event.request.body : undefined,
				// @ts-ignore
				duplex: 'half'
			};

			const response = await fetch(url, options);

			return new Response(response.body, {
				status: response.status,
				statusText: response.statusText,
				headers: response.headers
			});
		} catch (err) {
			console.error('Proxy Error:', err);
			return new Response('Proxy Error', { status: 502 });
		}
	}

	return resolve(event);
};
