import { createBrowserClient } from '@supabase/ssr'

export function createClient() {
  let jwtValue = '';
  if (typeof document !== 'undefined') {
    const match = document.cookie.match(/(?:^|; )jwt=([^;]*)/);
    if (match) {
      jwtValue = match[1];
    }
  }

  const options: { global?: { headers: Record<string, string> } } = {};
  if (jwtValue) {
    options.global = {
      headers: {
        Authorization: `Bearer ${jwtValue}`
      }
    };
  }

  return createBrowserClient(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!,
    options
  )
}
