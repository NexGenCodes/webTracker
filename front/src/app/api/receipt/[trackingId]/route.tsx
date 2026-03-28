import { ImageResponse } from 'next/og';
import { NextRequest } from 'next/server';

export const runtime = 'edge';

interface ShipmentPayload {
  tracking_id?: string;
  status?: string;
  origin?: string;
  destination?: string;
  sender_name?: string;
  recipient_name?: string;
  recipient_address?: string;
  weight?: number;
  cargo_type?: string;
  // sqlc nullable wrappers
  Status?: { String: string; Valid: boolean };
  Origin?: { String: string; Valid: boolean };
  Destination?: { String: string; Valid: boolean };
  SenderName?: { String: string; Valid: boolean };
  RecipientName?: { String: string; Valid: boolean };
  RecipientAddress?: { String: string; Valid: boolean };
  Weight?: { Float64: number; Valid: boolean };
  CargoType?: { String: string; Valid: boolean };
}

/** Unwrap sqlc nullable wrappers or plain values */
function str(val: unknown, fallback: string): string {
  if (typeof val === 'string' && val) return val;
  if (val && typeof val === 'object' && 'Valid' in (val as Record<string, unknown>)) {
    const wrapper = val as { Valid: boolean; String?: string };
    if (wrapper.Valid && wrapper.String) return wrapper.String;
  }
  return fallback;
}

function num(val: unknown, fallback: number): number {
  if (typeof val === 'number') return val;
  if (val && typeof val === 'object' && 'Valid' in (val as Record<string, unknown>)) {
    const wrapper = val as { Valid: boolean; Float64?: number };
    if (wrapper.Valid && typeof wrapper.Float64 === 'number') return wrapper.Float64;
  }
  return fallback;
}

export async function GET(
  req: NextRequest,
  { params }: { params: Promise<{ trackingId: string }> }
) {
  const { trackingId } = await params;
  const { searchParams } = new URL(req.url);

  // Fetch real shipment data from the Go backend
  let shipment: ShipmentPayload | null = null;
  try {
    const backendUrl = process.env.BACKEND_URL || 'http://localhost:5000';
    const headers: Record<string, string> = {};
    const apiKey = process.env.API_SECRET_KEY;
    if (apiKey) headers['X-API-Key'] = apiKey;

    const res = await fetch(`${backendUrl}/api/track/${trackingId}`, {
      headers,
      cache: 'no-store',
    });
    if (res.ok) {
      shipment = (await res.json()) as ShipmentPayload;
    }
  } catch {
    // Silently fall through to query-param / defaults
  }

  // Priority: backend data → query param override → fallback

  const origin  = searchParams.get('origin')   || str(shipment?.origin ?? shipment?.Origin, 'TBD');
  const dest    = searchParams.get('dest')     || str(shipment?.destination ?? shipment?.Destination, 'TBD');
  const sender  = searchParams.get('sender')   || str(shipment?.sender_name ?? shipment?.SenderName, '---');
  const receiver = searchParams.get('receiver') || str(shipment?.recipient_name ?? shipment?.RecipientName, '---');
  const rawWeight = num(shipment?.weight ?? shipment?.Weight, 0);
  const weight  = searchParams.get('weight')   || (rawWeight > 0 ? `${rawWeight.toFixed(2)} KGS` : '0.00 KGS');
  const content = searchParams.get('content')  || str(shipment?.cargo_type ?? shipment?.CargoType, 'CONSIGNMENT');
  const today = new Date().toLocaleDateString('en-US', {
    year: 'numeric', month: 'short', day: 'numeric'
  });

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
          overflow: 'hidden',
        }}
      >
        {/* Noise Pattern */}
        <div style={{ position: 'absolute', top: 0, left: 0, right: 0, bottom: 0, opacity: 0.03, backgroundImage: 'radial-gradient(#000 1px, transparent 1px)', backgroundSize: '10px 10px' }} />
        
        {/* Guilloche Pattern Simulation */}
        <div style={{ position: 'absolute', top: '350px', left: 0, width: '1200px', height: '150px', opacity: 0.05, display: 'flex', flexDirection: 'column' }}>
          {[...Array(6)].map((_, i) => (
            <div key={i} style={{ width: '100%', height: '2px', backgroundColor: '#8b0000', marginBottom: '20px', borderRadius: '50%', transform: `scaleY(${0.5 + i * 0.1})` }} />
          ))}
        </div>

        {/* Fold Lines */}
        <div style={{ position: 'absolute', top: '210px', left: 0, width: '100%', height: '1px', backgroundColor: 'rgba(0,0,0,0.05)' }} />
        <div style={{ position: 'absolute', top: '420px', left: 0, width: '100%', height: '1px', backgroundColor: 'rgba(0,0,0,0.05)' }} />

        {/* ORIGINAL Watermark */}
        <div style={{ position: 'absolute', top: '50%', left: '50%', transform: 'translate(-50%, -50%) rotate(-30deg)', color: 'rgba(139, 0, 0, 0.04)', fontSize: '180px', fontWeight: 900, pointerEvents: 'none' }}>
          ORIGINAL
        </div>

        {/* Header Section */}
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', width: '100%', marginBottom: '20px' }}>
          <h1 style={{ color: '#8b0000', fontSize: '64px', margin: '0 0 10px 0', textTransform: 'uppercase', fontWeight: 900, fontStyle: 'italic' }}>
            Airway Bill
          </h1>
          <div style={{ width: '600px', height: '32px', backgroundColor: '#8b0000', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <span style={{ color: 'white', fontSize: '12px', fontWeight: 800, letterSpacing: '0.4em' }}>
              INTERNATIONAL SPECIAL DELIVERY SERVICE
            </span>
          </div>
          <div style={{ display: 'flex', justifyContent: 'space-between', width: '100%', padding: '0 60px', marginTop: '10px' }}>
            <span style={{ fontSize: '18px', fontWeight: 900, color: 'rgba(0,0,0,0.6)' }}>{trackingId}</span>
            <span style={{ fontSize: '28px', fontWeight: 900, color: '#cc0000' }}>№ 00{trackingId}</span>
          </div>
        </div>

        {/* Main Grid */}
        <div style={{ display: 'flex', flexWrap: 'wrap', border: '2px solid black', width: '100%', backgroundColor: 'rgba(255,255,255,0.3)' }}>
          {/* Row 1 */}
          <div style={{ display: 'flex', width: '100%', borderBottom: '2px solid black', height: '70px' }}>
            <div style={{ width: '30%', padding: '10px 20px', borderRight: '2px solid black', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
              <span style={{ fontSize: '9px', color: '#464646', fontWeight: 800, opacity: 0.6 }}>DESTINATION</span>
              <span style={{ fontSize: '22px', color: '#2d2d2d', fontWeight: 900 }}>{dest.toUpperCase()}</span>
            </div>
            <div style={{ width: '22%', borderRight: '2px solid black', display: 'flex', flexDirection: 'column' }}>
               <div style={{ height: '24px', backgroundColor: 'black', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'white', fontSize: '10px', fontWeight: 900 }}>SERVICE</div>
               <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px' }}>
                  <span style={{ fontSize: '11px', fontWeight: 900, opacity: 0.4 }}>DIPLOMATIC</span>
                  <div style={{ width: '20px', height: '20px', border: '1px solid black', display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 900, fontSize: '12px' }}>X</div>
               </div>
            </div>
            <div style={{ width: '22%', borderRight: '2px solid black', display: 'flex', flexDirection: 'column' }}>
               <div style={{ height: '24px', backgroundColor: 'black', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'white', fontSize: '10px', fontWeight: 900 }}>PAYMENT</div>
               <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '8px' }}>
                  <span style={{ fontSize: '11px', fontWeight: 900, opacity: 0.4 }}>ACCOUNT</span>
                  <div style={{ width: '20px', height: '20px', border: '1px solid black', display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 900, fontSize: '12px' }}>X</div>
               </div>
            </div>
            <div style={{ width: '26%', padding: '10px 20px', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
              <span style={{ fontSize: '9px', color: '#464646', fontWeight: 800, opacity: 0.6 }}>ORIGIN</span>
              <span style={{ fontSize: '22px', color: '#2d2d2d', fontWeight: 900 }}>{origin.toUpperCase()}</span>
            </div>
          </div>
          
          {/* Row 2 (Details) */}
          <div style={{ display: 'flex', width: '100%', borderBottom: '2px solid black', height: '120px' }}>
            <div style={{ width: '52%', padding: '15px 25px', borderRight: '2px solid black', display: 'flex', flexDirection: 'column' }}>
              <span style={{ fontSize: '9px', color: '#464646', fontWeight: 800, opacity: 0.6, marginBottom: '5px' }}>CONSIGNEE (RECEIVER)</span>
              <span style={{ fontSize: '28px', color: '#2d2d2d', fontWeight: 900 }}>{receiver.toUpperCase()}</span>
              <span style={{ fontSize: '14px', fontStyle: 'italic', fontWeight: 700, color: 'rgba(0,0,0,0.6)' }}>SECURE DESTINATION AUTH REQUIRED</span>
            </div>
            <div style={{ width: '48%', padding: '15px 25px', display: 'flex', flexDirection: 'column' }}>
              <span style={{ fontSize: '9px', color: '#464646', fontWeight: 800, opacity: 0.6, marginBottom: '5px' }}>CONSIGNOR (SENDER)</span>
              <span style={{ fontSize: '28px', color: '#2d2d2d', fontWeight: 900 }}>{sender.toUpperCase()}</span>
            </div>
          </div>

          {/* Row 3 (Cargo) */}
          <div style={{ display: 'flex', width: '100%', height: '70px' }}>
            <div style={{ width: '30%', padding: '10px 20px', borderRight: '2px solid black', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
              <span style={{ fontSize: '9px', color: '#464646', fontWeight: 800, opacity: 0.6 }}>DESCRIPTION</span>
              <span style={{ fontSize: '20px', color: '#2d2d2d', fontWeight: 900 }}>{content.toUpperCase()}</span>
            </div>
            <div style={{ width: '22%', padding: '10px 20px', borderRight: '2px solid black', display: 'flex', flexDirection: 'column', justifyContent: 'center', backgroundColor: 'rgba(139, 0, 0, 0.05)' }}>
              <span style={{ fontSize: '9px', color: '#464646', fontWeight: 800, opacity: 0.6 }}>GROSS WT</span>
              <span style={{ fontSize: '24px', color: '#8b0000', fontWeight: 900 }}>{weight.toUpperCase()}</span>
            </div>
            <div style={{ width: '22%', padding: '10px 20px', borderRight: '2px solid black', display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
              <span style={{ fontSize: '9px', color: '#464646', fontWeight: 800, opacity: 0.6 }}>DEP DATE</span>
              <span style={{ fontSize: '20px', color: '#2d2d2d', fontWeight: 900 }}>{today}</span>
            </div>
            <div style={{ width: '26%', padding: '10px', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center' }}>
               <div style={{ width: '100%', backgroundColor: '#cc0000', color: 'white', fontSize: '9px', fontWeight: 900, textAlign: 'center', padding: '2px 0' }}>! CONFIDENTIAL !</div>
               <span style={{ fontSize: '8px', fontStyle: 'italic', fontWeight: 900, textAlign: 'center', opacity: 0.6 }}>RESTRICTED ACCESS DIPLOMATIC CARGO</span>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div style={{ display: 'flex', marginTop: 'auto', justifyContent: 'space-between', alignItems: 'flex-end', width: '100%' }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '15px' }}>
            <div style={{ display: 'flex', gap: '4px', alignItems: 'flex-end' }}>
              {[15, 8, 22, 6, 18, 10, 28, 6, 15, 8, 12, 18, 24, 6].map((h, i) => (
                <div key={i} style={{ width: '4px', height: `${h * 2}px`, backgroundColor: '#2d2d2d' }} />
              ))}
            </div>
            <div style={{ borderTop: '2px solid black', width: '300px', display: 'flex', flexDirection: 'column', paddingTop: '5px' }}>
               <span style={{ fontSize: '24px', fontStyle: 'italic', color: '#00008b', fontWeight: 900 }}>{sender}</span>
               <span style={{ fontSize: '9px', fontWeight: 900, opacity: 0.4, textTransform: 'uppercase' }}>Authorized Courier Initials</span>
            </div>
          </div>

          {/* Secure Stamp */}
          <div style={{ 
            border: '4px solid rgba(139, 0, 0, 0.25)', 
            borderRadius: '50%', 
            width: '120px', 
            height: '120px', 
            display: 'flex', 
            alignItems: 'center', 
            justifyContent: 'center', 
            transform: 'rotate(-15deg)', 
            color: 'rgba(139, 0, 0, 0.3)', 
            fontWeight: 900, 
            fontSize: '11px', 
            textAlign: 'center',
            position: 'relative',
            marginRight: '60px'
          }}>
            <div style={{ border: '1px dashed rgba(139, 0, 0, 0.2)', borderRadius: '50%', width: '100px', height: '100px', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              VERIFIED<br/>& SECURED<br/>DIPLOMATIC
            </div>
          </div>
          
          <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', gap: '5px' }}>
             <div style={{ display: 'flex', flexWrap: 'wrap', width: '60px', height: '60px', gap: '1px' }}>
                {[...Array(25)].map((_, i) => (
                  <div key={i} style={{ width: '11px', height: '11px', backgroundColor: Math.random() > 0.4 ? 'black' : 'transparent' }} />
                ))}
             </div>
             <span style={{ fontSize: '9px', backgroundColor: 'rgba(0,0,0,0.05)', padding: '2px 6px', borderRadius: '4px', fontWeight: 900, color: 'rgba(0,0,0,0.4)' }}>WTB-SECURE-V1.2</span>
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
