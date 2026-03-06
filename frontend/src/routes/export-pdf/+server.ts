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
        // Regex explanation:
        // (src|href)     -> Capture group 1: attribute name
        // \s*=\s*        -> Handle potential spaces around =
        // ["']           -> Match opening quote (either " or ')
        // \/             -> Match the first slash (root relative)
        // (?!\/)         -> Negative lookahead: ensure the next char is NOT a slash (avoid // protocol relative)
        // ([^"']+)       -> Capture group 2: the rest of the path until the closing quote
        // ["']           -> Match closing quote
        processedHtml = processedHtml.replace(/(src|href)\s*=\s*["']\/([^\/][^"']*)["']/g, '$1="http://frontend:3000/$2"');

        const formData = new FormData();
        
        // Send the pure, native HTML to Gotenberg
        formData.append('files', new Blob([processedHtml], { type: 'text/html' }), 'index.html');
        
        // Zero margins from Gotenberg so the DOM layout dictates the bounds perfectly
        formData.append('marginTop', '0');
        formData.append('marginBottom', '0');
        formData.append('marginLeft', '0');
        formData.append('marginRight', '0');
        formData.append('preferCssPageSize', 'true');
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
