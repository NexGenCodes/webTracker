'use server';

import { logger } from '@/lib/logger';

export interface ContactFormData {
    name: string;
    email: string;
    message: string;
}

export interface ContactResult {
    success: boolean;
    error?: string;
}

/**
 * Submit contact form message (validates and logs, does not store)
 */
export async function submitContactMessage(data: ContactFormData): Promise<ContactResult> {
    try {
        // Trim all inputs
        const name = data.name?.trim() || '';
        const email = data.email?.trim().toLowerCase() || '';
        const message = data.message?.trim() || '';

        // 1. Check required fields
        if (!name || !email || !message) {
            return { success: false, error: 'All fields are required' };
        }

        // 2. Name validation
        if (name.length < 2) {
            return { success: false, error: 'Name must be at least 2 characters' };
        }
        if (name.length > 100) {
            return { success: false, error: 'Name must be less than 100 characters' };
        }
        if (!/^[a-zA-Z\s'-]+$/.test(name)) {
            return { success: false, error: 'Name can only contain letters, spaces, hyphens, and apostrophes' };
        }

        // 3. Email validation
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(email)) {
            return { success: false, error: 'Please enter a valid email address' };
        }
        if (email.length > 255) {
            return { success: false, error: 'Email address is too long' };
        }

        // 4. Message validation
        if (message.length < 10) {
            return { success: false, error: 'Message must be at least 10 characters' };
        }
        if (message.length > 2000) {
            return { success: false, error: 'Message must be less than 2000 characters' };
        }

        // 5. Check for suspicious content (basic XSS protection)
        const suspiciousPatterns = /<script|javascript:|onerror=|onclick=/i;
        if (suspiciousPatterns.test(name) || suspiciousPatterns.test(message)) {
            return { success: false, error: 'Invalid content detected' };
        }

        // Log the contact message (but don't store it)
        logger.info('Contact message received (not stored)', {
            name,
            email,
            messageLength: message.length
        });

        // Simulate success without storing
        return { success: true };
    } catch (error) {
        logger.error('[ContactAction] Error processing contact message', error);
        return { success: false, error: 'Failed to send message. Please try again.' };
    }
}
