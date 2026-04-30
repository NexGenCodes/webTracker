import { NextResponse } from 'next/server';
import { paystackRequest } from '@/lib/paystack';


export async function POST(request: Request) {
    try {
        const companyId = request.headers.get('x-company-id');
        if (!companyId) {
            return NextResponse.json({ error: 'Company ID is required' }, { status: 401 });
        }

        const body = await request.json();
        const { planId, amount, email } = body;

        if (!planId || !amount || !email) {
            return NextResponse.json({ error: 'Missing required fields' }, { status: 400 });
        }

        // Initialize Paystack Transaction
        // For subscriptions, you would ideally pass a `plan` code here.
        // For now, we initialize a standard transaction and use `metadata` to track the upgrade.
        const response = await paystackRequest<{ authorization_url: string; access_code: string; reference: string }>('/transaction/initialize', {
            method: 'POST',
            body: JSON.stringify({
                email,
                amount: Math.round(amount * 100), // Convert NGN to Kobo
                callback_url: `${process.env.NEXT_PUBLIC_APP_URL || 'http://localhost:3000'}/dashboard/billing?status=success`,
                metadata: {
                    company_id: companyId,
                    plan_type: planId,
                    action: 'subscription_upgrade'
                }
            })
        });

        return NextResponse.json({ data: response.data });
    } catch (error: unknown) {
        console.error('Paystack Init Error:', error);
        return NextResponse.json({ error: error instanceof Error ? error.message : 'Failed to initialize payment' }, { status: 500 });
    }
}
