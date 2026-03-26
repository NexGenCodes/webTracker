import { ImageResponse } from 'next/og';
import { NextRequest } from 'next/server';

export const runtime = 'edge';

export async function GET(
  req: NextRequest,
  { params }: { params: { trackingId: string } }
) {
  const { trackingId } = params;

  // In a real app, you'd fetch shipment data here.
  // For now, we'll parse status from query params or use defaults.
  const { searchParams } = new URL(req.url);
  const status = searchParams.get('status') || 'PENDING';
  const origin = searchParams.get('origin') || 'TBD';
  const dest = searchParams.get('dest') || 'TBD';
  const sender = searchParams.get('sender') || '---';
  const receiver = searchParams.get('receiver') || '---';
  const weight = searchParams.get('weight') || '0.00 KGS';
  const content = searchParams.get('content') || 'CONSIGNMENT';

  return new ImageResponse(
    (
      <div
        style={{
          height: '100%',
          width: '100%',
          display: 'flex',
          flexDirection: 'column',
          backgroundColor: '#f4f2eb',
          fontFamily: 'serif',
          padding: '40px',
          position: 'relative',
        }}
      >
        {/* Background Noise/Texture Simulation */}
        <div style={{ position: 'absolute', top: 0, left: 0, right: 0, bottom: 0, opacity: 0.03, backgroundImage: 'radial-gradient(#000 1px, transparent 1px)', backgroundSize: '10px 10px' }} />
        
        {/* ORIGINAL Watermark */}
        <div style={{ position: 'absolute', top: '50%', left: '50%', transform: 'translate(-50%, -50%) rotate(-30deg)', color: 'rgba(139, 0, 0, 0.05)', fontSize: '200px', fontWeight: 900, pointerEvents: 'none' }}>
          ORIGINAL
        </div>

        {/* Header */}
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', width: '100%', marginBottom: '15px' }}>
          <h1 style={{ color: '#8b0000', fontSize: '50px', margin: '0 0 5px 0', textTransform: 'uppercase', fontWeight: 900 }}>
            Airway Bill
          </h1>
          <div style={{ width: '60%', height: '24px', backgroundColor: '#8b0000', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <span style={{ color: 'white', fontSize: '11px', fontWeight: 800, letterSpacing: '0.2em' }}>
              INTERNATIONAL SPECIAL DELIVERY SERVICE
            </span>
          </div>
          <span style={{ color: '#2d2d2d', fontSize: '16px', fontWeight: 700, marginTop: '8px' }}>{trackingId}</span>
        </div>

        {/* Info Grid */}
        <div style={{ display: 'flex', flexWrap: 'wrap', border: '1px solid black', width: '100%' }}>
          {/* Row 1 */}
          <div style={{ display: 'flex', width: '100%', borderBottom: '1px solid black' }}>
            <div style={{ width: '50%', padding: '10px', borderRight: '1px solid black', display: 'flex', flexDirection: 'column' }}>
              <span style={{ fontSize: '10px', color: '#464646', fontWeight: 800 }}>DESTINATION</span>
              <span style={{ fontSize: '18px', color: '#2d2d2d', fontWeight: 900 }}>{dest.toUpperCase()}</span>
            </div>
            <div style={{ width: '50%', padding: '10px', display: 'flex', flexDirection: 'column' }}>
              <span style={{ fontSize: '10px', color: '#464646', fontWeight: 800 }}>ORIGIN</span>
              <span style={{ fontSize: '18px', color: '#2d2d2d', fontWeight: 900 }}>{origin.toUpperCase()}</span>
            </div>
          </div>
          
          {/* Row 2 */}
          <div style={{ display: 'flex', width: '100%', borderBottom: '1px solid black' }}>
            <div style={{ width: '50%', padding: '10px', borderRight: '1px solid black', display: 'flex', flexDirection: 'column' }}>
              <span style={{ fontSize: '10px', color: '#464646', fontWeight: 800 }}>RECEIVER</span>
              <span style={{ fontSize: '18px', color: '#2d2d2d', fontWeight: 900 }}>{receiver.toUpperCase()}</span>
            </div>
            <div style={{ width: '50%', padding: '10px', display: 'flex', flexDirection: 'column' }}>
              <span style={{ fontSize: '10px', color: '#464646', fontWeight: 800 }}>SENDER</span>
              <span style={{ fontSize: '18px', color: '#2d2d2d', fontWeight: 900 }}>{sender.toUpperCase()}</span>
            </div>
          </div>

          {/* Row 3 */}
          <div style={{ display: 'flex', width: '100%', borderBottom: '1px solid black' }}>
            <div style={{ width: '50%', padding: '10px', borderRight: '1px solid black', display: 'flex', flexDirection: 'column' }}>
              <span style={{ fontSize: '10px', color: '#464646', fontWeight: 800 }}>CONTENT / TYPE</span>
              <span style={{ fontSize: '18px', color: '#2d2d2d', fontWeight: 900 }}>{content.toUpperCase()}</span>
            </div>
            <div style={{ width: '50%', padding: '10px', display: 'flex', flexDirection: 'column' }}>
              <span style={{ fontSize: '10px', color: '#464646', fontWeight: 800 }}>WEIGHT</span>
              <span style={{ fontSize: '18px', color: '#2d2d2d', fontWeight: 900 }}>{weight.toUpperCase()}</span>
            </div>
          </div>
          
          {/* Row 4 (Status) */}
          <div style={{ display: 'flex', width: '100%' }}>
            <div style={{ width: '100%', padding: '10px', display: 'flex', flexDirection: 'column', alignItems: 'center', backgroundColor: 'rgba(139, 0, 0, 0.03)' }}>
              <span style={{ fontSize: '10px', color: '#464646', fontWeight: 800 }}>CURRENT TRACKING STATUS</span>
              <span style={{ fontSize: '22px', color: '#8b0000', fontWeight: 900 }}>{status.replace(/_/g, ' ').toUpperCase()}</span>
            </div>
          </div>
        </div>

        {/* Footer Section */}
        <div style={{ display: 'flex', marginTop: 'auto', justifyContent: 'space-between', alignItems: 'flex-end', width: '100%' }}>
          {/* Barcode Simulation */}
          <div style={{ display: 'flex', gap: '2px' }}>
            {[10, 4, 15, 6, 12, 8, 20, 4, 10, 6, 8, 12, 18, 4, 10, 6, 12, 4].map((h, i) => (
              <div key={i} style={{ width: '3px', height: `${h * 3}px`, backgroundColor: '#2d2d2d' }} />
            ))}
          </div>

          {/* Approved Stamp */}
          <div style={{ border: '3px solid rgba(139, 0, 0, 0.4)', borderRadius: '50%', width: '110px', height: '110px', display: 'flex', alignItems: 'center', justifyContent: 'center', transform: 'rotate(-15deg)', color: 'rgba(139, 0, 0, 0.4)', fontWeight: 900, fontSize: '14px', textAlign: 'center' }}>
            SECURED<br/>DIPLOMATIC
          </div>
          
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end' }}>
             <span style={{ fontSize: '8px', color: 'rgba(0,0,0,0.4)', marginBottom: '3px' }}>OFFICIAL FREIGHT MANIFEST v1.1</span>
             <div style={{ width: '150px', height: '1px', backgroundColor: 'black' }} />
             <span style={{ fontSize: '12px', fontFamily: 'cursive', color: '#00008b', marginTop: '3px' }}>{origin.split(',')[0]} Transit Auth</span>
          </div>
        </div>
      </div>
    ),
    {
      width: 1200,
      height: 630,
    }
  );
}
