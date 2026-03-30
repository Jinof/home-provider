import { sqliteTable, text, integer, index } from 'drizzle-orm/sqlite-core';

export const providers = sqliteTable('providers', {
  id: text('id').primaryKey(),
  name: text('name').notNull().unique(),
  providerType: text('provider_type', { enum: ['openai', 'anthropic', 'google', 'glm', 'ollama'] }).notNull(),
  apiEndpoint: text('api_endpoint').notNull(),
  apiKeyEncrypted: text('api_key_encrypted').notNull(),
  models: text('models', { mode: 'json' }).notNull().$type<string[]>(),
  isActive: integer('is_active', { mode: 'boolean' }).notNull().default(true),
  createdAt: integer('created_at', { mode: 'timestamp' }).notNull(),
  updatedAt: integer('updated_at', { mode: 'timestamp' }).notNull(),
});

export const apiKeys = sqliteTable('api_keys', {
  id: text('id').primaryKey(),
  keyHash: text('key_hash').notNull().unique(),
  keyPrefix: text('key_prefix').notNull(),
  name: text('name').notNull(),
  allowedProviders: text('allowed_providers', { mode: 'json' }).$type<string[] | null>(),
  requestLimit: integer('request_limit'),
  expiresAt: integer('expires_at', { mode: 'timestamp' }),
  isActive: integer('is_active', { mode: 'boolean' }).notNull().default(true),
  createdAt: integer('created_at', { mode: 'timestamp' }).notNull(),
}, (table) => ({
  keyHashIdx: index('idx_api_keys_key_hash').on(table.keyHash),
}));

export const usageLogs = sqliteTable('usage_logs', {
  id: text('id').primaryKey(),
  apiKeyId: text('api_key_id').notNull().references(() => apiKeys.id),
  provider: text('provider').notNull(),
  model: text('model').notNull(),
  inputTokens: integer('input_tokens').notNull().default(0),
  outputTokens: integer('output_tokens').notNull().default(0),
  latencyMs: integer('latency_ms').notNull(),
  statusCode: integer('status_code').notNull(),
  createdAt: integer('created_at', { mode: 'timestamp' }).notNull(),
}, (table) => ({
  apiKeyIdIdx: index('idx_usage_logs_api_key_id').on(table.apiKeyId),
  createdAtIdx: index('idx_usage_logs_created_at').on(table.createdAt),
}));