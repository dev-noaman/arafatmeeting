import { createClient } from '@insforge/sdk';
import type { InsForgeClient, UserSchema } from '@insforge/sdk';

export const insforge: InsForgeClient = createClient({
  baseUrl: import.meta.env.VITE_INSFORGE_URL,
  anonKey: import.meta.env.VITE_INSFORGE_ANON_KEY,
});

export function getAccessToken(): string | null {
  const headers = insforge.getHttpClient().getHeaders();
  const auth = headers['Authorization'] || headers['authorization'];
  if (!auth) return null;
  return auth.replace('Bearer ', '');
}

export type { UserSchema };
