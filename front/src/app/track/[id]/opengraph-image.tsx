import { ImageResponse } from 'next/og';

export const alt = 'Shipment Tracking';
export const size = {
  width: 1200,
  height: 630,
};

export const contentType = 'image/png';

/**
 * Lightweight fetch directly to Go backend, avoiding server-action imports
 * that pull in next-auth / Sentry / logger (all Node-only, breaks edge).
 */
async function fetchTrackingData(id: string) {
  const baseUrl = process.env.BACKEND_URL || 'http://localhost:5000';
  const headers: Record<string, string> = {};
  if (process.env.API_SECRET_KEY) headers['X-API-Key'] = process.env.API_SECRET_KEY;

  try {
    const res = await fetch(`${baseUrl}/api/track/${id}`, {
      headers,
      next: { revalidate: 60 },
    });
    if (!res.ok) return null;
    return await res.json();
  } catch {
    return null;
  }
}

export default async function Image({ params }: { params: { id: string } }) {
  const id = params.id.toUpperCase();
  const data = await fetchTrackingData(id);

  if (!data) {
    return new ImageResponse(
      (
        <div
          style={{
            fontSize: 48,
            background: '#09090b',
            width: '100%',
            height: '100%',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'white',
            flexDirection: 'column',
          }}
        >
          <div style={{ color: '#3b82f6', fontWeight: 'bold', marginBottom: 20 }}>WebTracker</div>
          <div>Shipment Not Found</div>
        </div>
      ),
      { ...size }
    );
  }

  const status = (data.status || '').replace(/_/g, ' ');
  const origin = data.origin || 'Global';
  const destination = data.destination || 'Undisclosed';

  return new ImageResponse(
    (
      <div
        style={{
          background: '#09090b',
          width: '100%',
          height: '100%',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'flex-start',
          justifyContent: 'center',
          padding: '80px',
          color: 'white',
          position: 'relative',
        }}
      >
        {/* Background Dot Grid */}
        <div style={{
          position: 'absolute',
          inset: 0,
          opacity: 0.1,
          backgroundImage: 'radial-gradient(#ffffff 1px, transparent 1px)',
          backgroundSize: '20px 20px',
        }} />

        <div style={{ display: 'flex', alignItems: 'center', marginBottom: 40 }}>
          <div style={{
            background: '#3b82f6',
            width: 80,
            height: 80,
            borderRadius: '20px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            marginRight: 24,
            fontSize: 40,
          }}>
            📦
          </div>
          <div style={{ display: 'flex', flexDirection: 'column' }}>
            <div style={{ fontSize: 24, color: '#3b82f6', fontWeight: 'bold', letterSpacing: '4px' }}>WEBTRACKER</div>
            <div style={{ fontSize: 48, fontWeight: '900', letterSpacing: '-2px' }}>MANIFEST {id}</div>
          </div>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', width: '100%' }}>
          <div style={{ fontSize: 20, color: '#a1a1aa', fontWeight: 'bold', marginBottom: 8, letterSpacing: '2px' }}>CURRENT STATUS</div>
          <div style={{ fontSize: 72, fontWeight: '900', color: '#ffffff', marginBottom: 40, textTransform: 'uppercase' }}>
            {status}
          </div>

          <div style={{ display: 'flex', gap: 60 }}>
            <div style={{ display: 'flex', flexDirection: 'column' }}>
              <div style={{ fontSize: 16, color: '#3b82f6', fontWeight: 'bold', marginBottom: 4 }}>ORIGIN</div>
              <div style={{ fontSize: 32, fontWeight: 'bold' }}>{origin}</div>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', fontSize: 32, opacity: 0.3 }}>➔</div>
            <div style={{ display: 'flex', flexDirection: 'column' }}>
              <div style={{ fontSize: 16, color: '#3b82f6', fontWeight: 'bold', marginBottom: 4 }}>DESTINATION</div>
              <div style={{ fontSize: 32, fontWeight: 'bold' }}>{destination}</div>
            </div>
          </div>
        </div>
        
        <div style={{
          position: 'absolute',
          bottom: 40,
          right: 40,
          fontSize: 16,
          color: '#3b82f6',
          opacity: 0.5,
          fontWeight: 'bold'
        }}>
          REAL-TIME TELEMETRY SYSTEM ACTIVE
        </div>
      </div>
    ),
    { ...size }
  );
}
