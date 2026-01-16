import { useState } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { AdminDashboard } from './presentation/pages/Admin';
import { TrackingSearch } from './presentation/components/TrackingSearch';
import { ZustandRepository } from './infrastructure/ZustandRepository';
import { ShipmentService } from './application/ShipmentService';
import { ShipmentMap } from './presentation/components/ShipmentMap';
import type { Shipment } from './domain/Shipment';
import { CONFIG } from './config';
import { Anchor } from 'lucide-react';
import { ThemeToggle } from './presentation/components/ThemeToggle';

// Initialize Service
const repo = new ZustandRepository();
const service = new ShipmentService(repo);

function App() {
  const [isAdmin, setIsAdmin] = useState(false);

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route
          path="/admin"
          element={
            isAdmin
              ? <AdminDashboard service={service} onLogout={() => setIsAdmin(false)} />
              : <LoginPage onLogin={() => setIsAdmin(true)} />
          }
        />
      </Routes>
    </BrowserRouter>
  );
}

// Sub-components for cleaner file (usually separate files)

const LoginPage = ({ onLogin }: { onLogin: () => void }) => {
  const [user, setUser] = useState('');
  const [pass, setPass] = useState('');
  const [error, setError] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (user === 'admin' && pass === 'password') {
      onLogin();
    } else {
      setError(true);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="glass-panel p-8 w-full max-w-md animate-fade-in">
        <h2 className="text-2xl font-bold mb-6 text-center text-heading">Admin Login</h2>
        <form onSubmit={handleSubmit} className="space-y-4">
          <input
            className="w-full p-3 bg-input-bg border border-input-border rounded-lg text-main outline-none focus:border-accent placeholder-muted"
            placeholder="Username"
            value={user} onChange={e => setUser(e.target.value)}
          />
          <input
            className="w-full p-3 bg-input-bg border border-input-border rounded-lg text-main outline-none focus:border-accent placeholder-muted"
            type="password"
            placeholder="Password"
            value={pass} onChange={e => setPass(e.target.value)}
          />
          {error && <p className="text-red-400 text-sm">Invalid credentials</p>}
          <button className="btn-primary w-full">Login</button>
        </form>
      </div>
    </div>
  );
};

const HomePage = () => {
  const [tracking, setTracking] = useState<Shipment | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSearch = async (id: string) => {
    setLoading(true);
    setError('');
    setTracking(null);
    try {
      const result = await service.getTracking(id);
      if (result) {
        setTracking(result);
      } else {
        setError('Shipment not found. Please check your tracking number.');
      }
    } catch (e) {
      setError('System error. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen p-4 flex flex-col items-center">
      {/* Header */}
      <header className="w-full max-w-6xl flex justify-between items-center py-6 mb-12">
        <div className="flex items-center gap-2 text-gradient font-bold text-xl">
          <Anchor className="text-accent" />
          {CONFIG.APP_NAME}
        </div>
        <div className="flex items-center gap-4">
          <ThemeToggle />
        </div>
      </header>

      {/* Main Content */}
      <main className="w-full max-w-4xl flex-1 flex flex-col items-center">
        <TrackingSearch onSearch={handleSearch} isLoading={loading} />

        {error && (
          <div className="mt-8 p-4 bg-red-500/10 border border-red-500/50 rounded-xl text-red-200 animate-fade-in">
            {error}
          </div>
        )}

        {tracking && (
          <div className="mt-12 w-full animate-fade-in">
            <div className="glass-panel p-8">
              <div className="flex justify-between items-start mb-6 border-b border-gray-700/50 pb-6">
                <div>
                  <p className="text-muted text-sm mb-1">Status</p>
                  <h2 className="text-3xl font-bold text-accent">{tracking.status.replace(/_/g, ' ')}</h2>
                </div>
                <div className="text-right">
                  <p className="text-muted text-sm mb-1">Destination</p>
                  <h3 className="text-xl font-semibold text-heading">{tracking.receiverCountry}</h3>
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                <div>
                  <h4 className="text-heading font-semibold mb-4">Shipment Details</h4>
                  <div className="space-y-3 text-sm text-main">
                    <p><span className="text-muted w-24 inline-block">Tracking ID:</span> {tracking.trackingNumber}</p>
                    <p><span className="text-muted w-24 inline-block">Receiver:</span> {tracking.receiverName}</p>
                    <p><span className="text-muted w-24 inline-block">From:</span> {tracking.senderName}</p>
                  </div>
                </div>

                <div className="border-l border-gray-700/50 pl-0 md:pl-8">
                  <h4 className="text-heading font-semibold mb-4">Latest Updates</h4>
                  <div className="space-y-6 relative ml-2">
                    {/* Timeline Line */}
                    <div className="absolute left-0 top-2 bottom-2 w-0.5 bg-input-border"></div>

                    {tracking.events.slice().reverse().map((event, i) => (
                      <div key={event.id} className="relative pl-6">
                        <div className={`absolute left-[-4px] top-1.5 w-2.5 h-2.5 rounded-full ${i === 0 ? 'bg-accent shadow-[0_0_10px_var(--color-accent)]' : 'bg-muted'}`}></div>
                        <p className="text-main font-medium">{event.status.replace(/_/g, ' ')}</p>
                        <p className="text-xs text-muted">{new Date(event.timestamp).toLocaleString()}</p>
                        <p className="text-sm text-main mt-1">{event.location}</p>
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              {/* Map Visualization */}
              <ShipmentMap locationName={tracking.events[tracking.events.length - 1].location} />
            </div>
          </div>
        )}
      </main>

      <footer className="w-full py-6 text-center text-gray-600 text-sm mt-auto">
        &copy; {new Date().getFullYear()} {CONFIG.APP_NAME} Logistics. Global tracking system.
      </footer>
    </div>
  );
};

export default App;
