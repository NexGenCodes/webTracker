import { createClient } from '@supabase/supabase-js';

const supabase = createClient(
  'https://ujfqzyixckoqbbvsakbg.supabase.co',
  'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InVqZnF6eWl4Y2tvcWJidnNha2JnIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc3Njc4MDk2NiwiZXhwIjoyMDkyMzU2OTY2fQ.9VBzEY5-Vb-kLNs2-j-R1uyYEKvI4wD2rXePmzdULVg',
  { db: { schema: 'public' } }
);

// Try to get JWT secret via raw SQL
const { data: jwtData, error: jwtErr } = await supabase.rpc('exec_sql', { 
  query: "SELECT current_setting('app.settings.jwt_secret', true) as secret" 
});
console.log('JWT via RPC:', jwtData, jwtErr?.message);

// Check RLS status on our tables
const { data: rlsData, error: rlsErr } = await supabase.rpc('exec_sql', {
  query: `SELECT tablename, rowsecurity FROM pg_tables WHERE schemaname = 'public' AND tablename IN ('companies', 'shipment')`
});
console.log('RLS status:', rlsData, rlsErr?.message);

// Check policies
const { data: polData, error: polErr } = await supabase.rpc('exec_sql', {
  query: `SELECT tablename, policyname, roles, cmd, qual FROM pg_policies WHERE tablename IN ('companies', 'shipment')`
});
console.log('Policies:', polData, polErr?.message);

// If RPC doesn't exist, try the SQL endpoint directly
const response = await fetch('https://ujfqzyixckoqbbvsakbg.supabase.co/rest/v1/rpc/exec_sql', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'apikey': 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InVqZnF6eWl4Y2tvcWJidnNha2JnIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc3Njc4MDk2NiwiZXhwIjoyMDkyMzU2OTY2fQ.9VBzEY5-Vb-kLNs2-j-R1uyYEKvI4wD2rXePmzdULVg',
    'Authorization': 'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZSIsInJlZiI6InVqZnF6eWl4Y2tvcWJidnNha2JnIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImlhdCI6MTc3Njc4MDk2NiwiZXhwIjoyMDkyMzU2OTY2fQ.9VBzEY5-Vb-kLNs2-j-R1uyYEKvI4wD2rXePmzdULVg',
  },
  body: JSON.stringify({ query: "SELECT current_setting('app.settings.jwt_secret', true)" })
});
console.log('Direct REST:', response.status, await response.text());

process.exit(0);
