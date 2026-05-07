import { cookies } from 'next/headers';
import { jwtVerify, importSPKI } from 'jose';

if (!process.env.JWT_PUBLIC_KEY) {
    throw new Error('FATAL: JWT_PUBLIC_KEY environment variable is not set.');
}

const JWT_PUBLIC_KEY_RAW = process.env.JWT_PUBLIC_KEY.replace(/\\n/g, '\n');

let publicKey: CryptoKey | null = null;
async function getPublicKey() {
  if (!publicKey) {
    publicKey = await importSPKI(JWT_PUBLIC_KEY_RAW, 'ES256');
  }
  return publicKey;
}

export interface SessionUser {
    company_id: string;
    company_name: string;
    email: string;
    plan_type: string;
    auth_status: string;
    role: string;
}

/**
 * Server-side auth check.
 * Verifies the backend JWT cookie using jose.
 */
export async function getServerSession(): Promise<{ authenticated: boolean; user?: SessionUser; token?: string }> {
    try {
        const cookieStore = await cookies();
        const jwt = cookieStore.get('jwt')?.value;
        
        if (!jwt) {
            return { authenticated: false };
        }

        const key = await getPublicKey();
        const { payload } = await jwtVerify(jwt, key);
        
        return { 
            authenticated: true,
            token: jwt,
            user: {
                company_id: payload.company_id as string,
                company_name: payload.company_name as string,
                email: payload.email as string,
                plan_type: payload.plan_type as string,
                auth_status: payload.auth_status as string,
                role: payload.role as string,
            }
        };
    } catch {
        return { authenticated: false };
    }
}
