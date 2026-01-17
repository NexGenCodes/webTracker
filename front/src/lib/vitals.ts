// Simple Server-Side Vitals Monitoring
// This mimics the 'Vitals' system added to the Go backend.

type VitalsMetric =
    | 'SHIPMENT_CREATED'
    | 'TRACKING_REQUESTED'
    | 'TRANSITION_DUE_CHECK'
    | 'ERROR_SERVER_ACTION';

class FrontendVitals {
    private static instance: FrontendVitals;
    private metrics: Record<string, number> = {};

    private constructor() { }

    public static getInstance(): FrontendVitals {
        if (!FrontendVitals.instance) {
            FrontendVitals.instance = new FrontendVitals();
        }
        return FrontendVitals.instance;
    }

    public track(metric: VitalsMetric) {
        this.metrics[metric] = (this.metrics[metric] || 0) + 1;

        // Log occasionally or under high load
        if (this.metrics[metric] % 5 === 0) {
            console.log(`[VITALS] ${metric}: ${this.metrics[metric]}`);
        }
    }

    public getSnapshot() {
        return { ...this.metrics };
    }
}

export const vitals = FrontendVitals.getInstance();
