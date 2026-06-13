import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},

			// Static SPA build: emitted to web/build, with index.html as the
			// fallback so the Go server can serve client-side routes.
			adapter: adapter({ fallback: 'index.html', strict: false })
		})
	],
	server: {
		proxy: {
			// Proxy API calls to the Go backend (which serves them under /api).
			'/api': {
				target: 'http://localhost:8080',
				changeOrigin: true
			}
		}
	}
});
