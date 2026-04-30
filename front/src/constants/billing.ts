export const BILLING_PLANS = [
    {
        id: 'starter',
        name: 'Starter',
        price: '₦10,000',
        interval: '/ mo',
        description: 'Perfect for independent dispatch riders and small couriers.',
        trial: '7 Days Free Trial',
        features: [
            '50 Shipments per month',
            'WhatsApp Bot Automation',
            'Web Tracking Portal',
            'Manual Shipment Entry',
            'Community Support'
        ],
        buttonText: 'Start Free Trial',
        popular: false,
    },
    {
        id: 'pro',
        name: 'Pro',
        price: '₦25,000',
        interval: '/ mo',
        description: 'For growing logistics companies that need to save time.',
        trial: null,
        features: [
            '250 Shipments per month',
            'WhatsApp Bot Automation',
            'AI PDF/Text Parser (Waybills)',
            'Bulk CSV Uploads',
            'Custom Branding (Logo & Colors)',
            'Priority Email Support'
        ],
        buttonText: 'Upgrade to Pro',
        popular: true,
    },
    {
        id: 'enterprise',
        name: 'Scale',
        price: '₦50,000',
        interval: '/ mo',
        description: 'High volume capacity for established freight forwarders.',
        trial: null,
        features: [
            '1,000 Shipments per month',
            'Everything in Pro',
            'API & Webhook Access',
            'Dedicated WhatsApp Number',
            '24/7 Phone & Engineering Support'
        ],
        buttonText: 'Contact Sales',
        popular: false,
    }
];
