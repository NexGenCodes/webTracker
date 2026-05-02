import { NextResponse, type NextRequest } from 'next/server'
import { jwtVerify, importSPKI } from 'jose'

const JWT_PUBLIC_KEY_RAW = process.env.NEXT_PUBLIC_JWT_PUBLIC_KEY?.replace(/\\n/g, '\n');
if (!JWT_PUBLIC_KEY_RAW) {
  console.error('CRITICAL: NEXT_PUBLIC_JWT_PUBLIC_KEY environment variable is not set. Auth will reject all tokens.');
}

let publicKey: CryptoKey | null = null;
async function getPublicKey() {
  if (!publicKey && JWT_PUBLIC_KEY_RAW) {
    publicKey = await importSPKI(JWT_PUBLIC_KEY_RAW, 'ES256');
  }
  return publicKey;
}

export async function middleware(request: NextRequest) {
  const jwt = request.cookies.get('jwt')?.value

  const isAuthPage = request.nextUrl.pathname.startsWith('/auth')
  const isProtectedPage = request.nextUrl.pathname.startsWith('/admin') || request.nextUrl.pathname.startsWith('/dashboard')

  let isValid = false
  if (jwt) {
    try {
      const key = await getPublicKey()
      if (key) {
        await jwtVerify(jwt, key)
        isValid = true
      }
    } catch {
      // Invalid JWT
      isValid = false
    }
  }

  // Authenticated users shouldn't see the auth page
  if (isAuthPage && isValid) {
    return NextResponse.redirect(new URL('/dashboard', request.url))
  }

  // Unauthenticated users can't access protected pages
  if (isProtectedPage && !isValid) {
    return NextResponse.redirect(new URL('/auth', request.url))
  }

  return NextResponse.next()
}

export const config = {
  matcher: [
    '/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
  ],
}
