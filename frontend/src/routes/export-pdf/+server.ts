import { error } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = async ({ request }) => {
    try {
        const { html, filename } = await request.json();

        if (!html) {
            throw error(400, 'HTML content is required');
        }

        // Replace relative URLs so Gotenberg can resolve them internally via the frontend container
        let processedHtml = html;
        processedHtml = processedHtml.replace(/src="\/([^"]+)"/g, 'src="http://frontend:3000/$1"');
        processedHtml = processedHtml.replace(/href="\/([^"]+)"/g, 'href="http://frontend:3000/$1"');

        const formData = new FormData();
        
        // Send the pure, native HTML to Gotenberg
        formData.append('files', new Blob([processedHtml], { type: 'text/html' }), 'index.html');
        
        // Use exact A4 size in inches to perfectly match Paged.js
        formData.append('paperWidth', '8.27');
        formData.append('paperHeight', '11.69');
        
        // We let Gotenberg use standard 'print' media type with native margins.
        formData.append('marginTop', '0');
        formData.append('marginBottom', '0');
        formData.append('marginLeft', '0');
        formData.append('marginRight', '0');
        formData.append('printBackground', 'true');
        
        // Wait long enough for fonts and base64 images to fully load (crucial for heavy base64 images)
        formData.append('waitDelay', '5s');

        // Set a long timeout for the Node fetch to Gotenberg (5 minutes)
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), 300000);

        try {
            // Call Gotenberg (internal docker network)
            const response = await fetch('http://gotenberg:3000/forms/chromium/convert/html', {
                method: 'POST',
                body: formData,
                signal: controller.signal
            });

            clearTimeout(timeoutId);

            if (!response.ok) {
                const errText = await response.text();
                console.error('Gotenberg Error:', errText);
                throw error(500, `Failed to generate PDF from Gotenberg: ${errText}`);
            }

            const pdfBuffer = await response.arrayBuffer();
            
            return new Response(pdfBuffer, {
                headers: {
                    'Content-Type': 'application/pdf',
                    'Content-Disposition': `attachment; filename="${filename || 'document'}.pdf"`
                }
            });
        } catch (fetchErr) {
            clearTimeout(timeoutId);
            console.error('Gotenberg Fetch Error:', fetchErr);
            throw error(500, 'Koneksi ke sistem pembuat PDF terputus (Timeout/Error). Silakan coba lagi.');
        }
    } catch (e) {
        console.error('Export PDF error:', e);
        throw error(500, 'Internal Server Error');
    }
};
