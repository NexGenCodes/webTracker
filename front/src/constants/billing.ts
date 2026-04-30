export const BILLING_PLANS = [
    {
        id: 'starter',
        nameKey: 'starterName',
        price: '₦14,900',
        intervalKey: 'monthlyInterval',
        descKey: 'starterDesc',
        trialKey: 'sevenDayTrial',
        features: [
            'feat_50_shipments',
            'feat_whatsapp',
            'feat_web_portal',
            'feat_manual_entry',
            'feat_community'
        ],
        btnKey: 'btnStartTrial',
        popular: false,
    },
    {
        id: 'pro',
        nameKey: 'proName',
        price: '₦59,900',
        intervalKey: 'monthlyInterval',
        descKey: 'proDesc',
        trial: null,
        features: [
            'feat_250_shipments',
            'feat_whatsapp',
            'feat_ai_parser',
            'feat_csv_upload',
            'feat_custom_branding',
            'feat_priority_support'
        ],
        btnKey: 'btnUpgradePro',
        popular: true,
    },
    {
        id: 'enterprise',
        nameKey: 'scaleName',
        price: '₦225,000',
        intervalKey: 'monthlyInterval',
        descKey: 'scaleDesc',
        trial: null,
        features: [
            'feat_1000_shipments',
            'feat_all_pro',
            'feat_api_webhook',
            'feat_dedicated_whatsapp',
            'feat_247_support'
        ],
        btnKey: 'btnContactSales',
        popular: false,
    }
];
