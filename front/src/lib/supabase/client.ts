import { createBrowserClient } from '@supabase/ssr'
import { SupabaseClient } from '@supabase/supabase-js'

let browserClient: SupabaseClient | null = null;
let lastJwtValue = '';

/**
 * Creates or returns the singleton Supabase browser client.
 * 
 * Since the JWT cookie is HttpOnly (not readable by JS), the token
 * must be passed from a server-rendered prop or fetched via a server action.
 * If no token is provided, falls back to the existing singleton.
 */
export function createClient(jwt?: string) {
  if (!jwt && browserClient) {
    return browserClient;
  }

  const jwtValue = jwt || '';

  // If we already have a client and the JWT hasn't changed, return the singleton
  if (browserClient && jwtValue === lastJwtValue) {
    return browserClient;
  }

  const options: { global?: { headers: Record<string, string> } } = {};
  if (jwtValue) {
    options.global = {
      headers: {
        Authorization: `Bearer ${jwtValue}`
      }
    };
  }

  browserClient = createBrowserClient(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!,
    options
  );
  lastJwtValue = jwtValue;

  return browserClient;
}
