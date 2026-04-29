import { NextResponse, type NextRequest } from 'next/server'
import { jwtVerify } from 'jose/jwt/verify'

const JWT_SECRET_RAW = process.env.JWT_SECRET;
if (!JWT_SECRET_RAW) {
  console.error('CRITICAL: JWT_SECRET environment variable is not set. Auth will reject all tokens.');
}
const JWT_SECRET = new TextEncoder().encode(JWT_SECRET_RAW || '');

export async function middleware(request: NextRequest) {
  const jwt = request.cookies.get('jwt')?.value

  const isAuthPage = request.nextUrl.pathname.startsWith('/auth')
  const isProtectedPage = request.nextUrl.pathname.startsWith('/admin') || request.nextUrl.pathname.startsWith('/dashboard')
  const isHomePage = request.nextUrl.pathname === '/'

  let isValid = false
  if (jwt) {
    try {
      await jwtVerify(jwt, JWT_SECRET)
      isValid = true
    } catch (e) {
      // Invalid JWT
      isValid = false
    }
  }

  // Authenticated users shouldn't see the auth or home pages
  if ((isAuthPage || isHomePage) && isValid) {
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
