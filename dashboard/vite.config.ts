import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig, loadEnv } from 'vite';

export default defineConfig(({ mode }) => {
	const env = loadEnv(mode, process.cwd(), '');

	const apiUrl = env.API_URL || 'http://localhost:8081';
	const webhookUrl = env.WEBHOOK_URL || 'http://localhost:8080';

	return {
		plugins: [sveltekit()],
		server: {
			proxy: {
				'/api': apiUrl,
				'/webhook': webhookUrl
			}
		}
	};
});
