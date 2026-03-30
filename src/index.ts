import { app } from './app.js';

const port = Number(process.env.PORT) || 3000;

app.listen({ port, host: '0.0.0.0' })
  .then((address) => console.log(`Server listening at ${address}`))
  .catch((err) => {
    console.error('Failed to start server:', err);
    process.exit(1);
  });