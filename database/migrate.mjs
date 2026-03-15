import postgres from 'postgres';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const migrationsDir = path.join(__dirname, 'migrations');

async function migrate() {
    const databaseUrl = process.env.DATABASE_URL;
    if (!databaseUrl) {
        console.error('DATABASE_URL environment variable is required');
        process.exit(1);
    }

    const sql = postgres(databaseUrl);

    try {
        const files = fs.readdirSync(migrationsDir).filter(f => f.endsWith('.sql')).sort();
        
        console.log(`🚀 Found ${files.length} migrations...`);

        for (const file of files) {
            console.log(`📝 Applying ${file}...`);
            const content = fs.readFileSync(path.join(migrationsDir, file), 'utf8');
            
            // Basic SQL execution. For a real production app, we'd have a migrations table.
            // But for this consolidation task, we'll just run them.
            await sql.unsafe(content);
            console.log(`✅ ${file} applied successfully.`);
        }

        console.log('🎉 All migrations finished!');
    } catch (err) {
        console.error('❌ Migration failed:', err);
        process.exit(1);
    } finally {
        await sql.end();
    }
}

migrate();
