import { createBrowserClient } from '@supabase/ssr'
import { SupabaseClient } from '@supabase/supabase-js'

let browserClient: SupabaseClient | null = null;
let lastJwtValue = '';

export function createClient() {
  let jwtValue = '';
  if (typeof document !== 'undefined') {
    const match = document.cookie.match(/(?:^|; )jwt=([^;]*)/);
    if (match) {
      jwtValue = match[1];
    }
  }

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
