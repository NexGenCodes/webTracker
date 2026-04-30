import { cookies } from 'next/headers';
import { jwtVerify } from 'jose';

const JWT_SECRET = new TextEncoder().encode(process.env.JWT_SECRET || 'fallback_secret_change_me');

export interface SessionUser {
    company_id: string;
    company_name: string;
    email: string;
    plan_type: string;
    auth_status: string;
}

/**
 * Server-side auth check.
 * Verifies the backend JWT cookie using jose.
 */
export async function getServerSession(): Promise<{ authenticated: boolean; user?: SessionUser }> {
    try {
        const cookieStore = await cookies();
        const jwt = cookieStore.get('jwt')?.value;
        
        if (!jwt) {
            return { authenticated: false };
        }

        const { payload } = await jwtVerify(jwt, JWT_SECRET);
        
        return { 
            authenticated: true,
            user: {
                company_id: payload.company_id as string,
                company_name: payload.company_name as string,
                email: payload.email as string,
                plan_type: payload.plan_type as string,
                auth_status: payload.auth_status as string,
            }
        };
    } catch {
        return { authenticated: false };
    }
}
