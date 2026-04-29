export const BILLING_PLANS = [
    {
        id: 'starter',
        name: 'Starter',
        price: '₦10,000',
        interval: '/ mo',
        description: 'Perfect for testing the waters. Comes with a 7-day free trial.',
        trial: '7 Days Free Trial',
        features: [
            '50 Shipments per month (Capped)',
            'Web Tracking Portal',
            'Basic Email Notifications',
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
        description: 'For growing logistics companies that need automated updates.',
        trial: null,
        features: [
            '500 Shipments per month',
            'WhatsApp Bot Automation',
            'Custom Branding (Logo & Colors)',
            'API Access',
            'Priority Email Support'
        ],
        buttonText: 'Upgrade to Pro',
        popular: true,
    },
    {
        id: 'enterprise',
        name: 'Enterprise',
        price: 'Custom',
        interval: '',
        description: 'Unlimited scale for massive operations.',
        trial: null,
        features: [
            'Unlimited Shipments',
            'Dedicated WhatsApp Number',
            'Custom Reporting & Analytics',
            '99.9% Uptime SLA',
            '24/7 Phone & Engineering Support'
        ],
        buttonText: 'Contact Sales',
        popular: false,
    }
];
