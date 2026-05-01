import { z } from 'zod';

export type AuthMode = 'signin' | 'register' | 'forgot-password' | 'reset-password';
export type RegisterStep = 'credentials' | 'otp';

export const signInSchema = z.object({
    email: z.string().email("Invalid email address"),
    password: z.string().min(1, "Password is required"),
});
export type SignInForm = z.infer<typeof signInSchema>;

export const registerSchema = z.object({
    companyName: z.string().min(2, "Company name must be at least 2 characters"),
    email: z.string().email("Invalid email address"),
    password: z.string().min(8, "Password must be at least 8 characters"),
    confirmPassword: z.string(),
    acceptTerms: z.boolean().refine(val => val === true, "You must accept the terms and privacy policy"),
}).refine((data) => data.password === data.confirmPassword, {
    message: "Passwords do not match",
    path: ["confirmPassword"],
});
export type RegisterForm = z.infer<typeof registerSchema>;

export const forgotPasswordSchema = z.object({
    email: z.string().email("Invalid email address"),
});
export type ForgotPasswordForm = z.infer<typeof forgotPasswordSchema>;

export const resetPasswordSchema = z.object({
    password: z.string().min(8, "New password must be at least 8 characters"),
    confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
    message: "Passwords do not match",
    path: ["confirmPassword"],
});
export type ResetPasswordForm = z.infer<typeof resetPasswordSchema>;
