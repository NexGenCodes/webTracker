import { useTransition } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { signInSchema, SignInForm } from '@/lib/validations/auth';
import { loginAction } from '@/app/actions/auth';
import { useMultiTenant } from '@/components/providers/MultiTenantProvider';

export function useSignIn(setError: (msg: string | null) => void) {
    const router = useRouter();
    const searchParams = useSearchParams();
    const [isPending, startTransition] = useTransition();
    const { refreshAuth } = useMultiTenant();

    const form = useForm<SignInForm>({ resolver: zodResolver(signInSchema) });

    const onSubmit = (data: SignInForm) => {
        setError(null);
        
        startTransition(async () => {
            const result = await loginAction(data);
            
            if (!result.success) {
                setError(result.error || 'Login failed. Please try again.');
                return;
            }

            await refreshAuth();
            
            const redirectUrl = searchParams.get('redirect') || searchParams.get('callbackUrl') || searchParams.get('returnUrl') || '/dashboard';
            window.location.href = redirectUrl;
        });
    };

    return { form, onSubmit, loading: isPending };
}
