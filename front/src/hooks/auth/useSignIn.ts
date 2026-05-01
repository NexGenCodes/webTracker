import { useTransition } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { signInSchema, SignInForm } from '@/lib/validations/auth';
import { loginAction } from '@/app/actions/auth';

export function useSignIn(setError: (msg: string | null) => void) {
    const router = useRouter();
    const searchParams = useSearchParams();
    const [isPending, startTransition] = useTransition();

    const form = useForm<SignInForm>({ resolver: zodResolver(signInSchema) });

    const onSubmit = (data: SignInForm) => {
        setError(null);
        
        startTransition(async () => {
            try {
                await loginAction(data);
                
                const redirectUrl = searchParams.get('redirect') || searchParams.get('callbackUrl') || searchParams.get('returnUrl') || '/dashboard';
                router.push(redirectUrl);
            } catch (err: unknown) {
                setError(err instanceof Error ? err.message : 'Network error. Please try again.');
            }
        });
    };

    return { form, onSubmit, loading: isPending };
}
