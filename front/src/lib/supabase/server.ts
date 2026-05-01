import { createServerClient } from '@supabase/ssr'
import { cookies } from 'next/headers'

export async function createClient() {
  const cookieStore = await cookies()
  const jwt = cookieStore.get('jwt')?.value

  type Cookie = { name: string; value: string; options?: unknown };
  
  const options: {
    cookies: {
      getAll: () => { name: string; value: string }[];
      setAll: (cookiesToSet: Cookie[]) => void;
    };
    global?: { headers: Record<string, string> };
  } = {
    cookies: {
      getAll() {
        return cookieStore.getAll()
      },
      setAll(cookiesToSet: Cookie[]) {
        try {
          cookiesToSet.forEach(({ name, value, options }) =>
            // @ts-expect-error - Supabase CookieOptions is compatible with Next.js ResponseCookie
            cookieStore.set(name, value, options)
          )
        } catch {
          // The `setAll` method was called from a Server Component.
          // This can be ignored if you have middleware refreshing
          // user sessions.
        }
      },
    },
  };

  if (jwt) {
    options.global = {
      headers: {
        Authorization: `Bearer ${jwt}`
      }
    };
  }

  return createServerClient(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!,
    options
  )
}
