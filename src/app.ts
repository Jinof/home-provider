import Fastify from 'fastify';
import cors from '@fastify/cors';

export const app = Fastify({
  logger: true,
});

await app.register(cors, { origin: true });

// Health check
app.get('/health', async () => ({ status: 'ok' }));