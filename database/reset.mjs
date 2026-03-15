import postgres from 'postgres';

async function reset() {
    const databaseUrl = process.env.DATABASE_URL;
    if (!databaseUrl) {
        console.error('DATABASE_URL environment variable is required');
        process.exit(1);
    }

    const sql = postgres(databaseUrl);

    try {
        console.log('🔄 Resetting database...');
        
        // Drop all tables in the public schema
        await sql.unsafe(`
            DROP TABLE IF EXISTS Shipment CASCADE;
            DROP TABLE IF EXISTS UserPreference CASCADE;
            DROP TABLE IF EXISTS GroupAuthority CASCADE;
            DROP TABLE IF EXISTS SystemConfig CASCADE;
            DROP TABLE IF EXISTS country_timezones CASCADE;

            DROP FUNCTION IF EXISTS generate_tracking_id() CASCADE;
            DROP FUNCTION IF EXISTS fn_shipment_auto_schedule() CASCADE;
            DROP FUNCTION IF EXISTS fn_process_status_transitions(TIMESTAMP) CASCADE;
            DROP FUNCTION IF EXISTS fn_prune_aged_shipments() CASCADE;
        `);
        
        console.log('✅ All tables dropped.');
        console.log('📝 Re-running migrations...');
        
        // Note: The migrate.mjs script will be run after this by the package.json script
    } catch (err) {
        console.error('❌ Reset failed:', err);
        process.exit(1);
    } finally {
        await sql.end();
    }
}

reset();
