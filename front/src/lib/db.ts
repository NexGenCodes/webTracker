import postgres from 'postgres';

const globalForDb = global as unknown as { conn: ReturnType<typeof postgres> | undefined };

const conn = globalForDb.conn ?? postgres(process.env.DATABASE_URL!, {
    ssl: 'require',
    max: 10,
    idle_timeout: 20,
    connect_timeout: 30,
});

if (process.env.NODE_ENV !== 'production') globalForDb.conn = conn;

export default conn;
