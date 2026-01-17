import { describe, it, expect, beforeEach } from 'vitest';
import { vitals } from '../vitals';

describe('Frontend Vitals', () => {
    it('should track metrics correctly', () => {
        vitals.track('SHIPMENT_CREATED');
        vitals.track('SHIPMENT_CREATED');

        const snapshot = vitals.getSnapshot();
        expect(snapshot['SHIPMENT_CREATED']).toBeGreaterThanOrEqual(2);
    });

    it('should return a record of all tracked metrics', () => {
        vitals.track('TRACKING_REQUESTED');
        const snapshot = vitals.getSnapshot();
        expect(snapshot).toHaveProperty('TRACKING_REQUESTED');
    });
});
