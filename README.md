# Airway Bill - Premium Logistics Tracker

A full-stack logistics tracking application built with Next.js, Prisma, and SQLite.

## Features

- **Deterministic Tracking IDs**: IDs generated from shipment data (format: `AWB-[UniqueHash]`).
- **Data Retention Policy**: Automatic anonymization of personal data (PII) upon delivery.
- **Admin Dashboard**: Secure interface to create shipments from email content.
- **Mobile-First Design**: Optimized for tracking on the go with glassmorphism aesthetics.

## Getting Started

### Prerequisites

- Node.js 18+
- npm

### Installation

1. Clone the repository
2. Install dependencies:

   ```bash
   npm install
   ```

3. Initialize the database:

   ```bash
   npx prisma generate
   npx prisma migrate dev --name init
   ```

4. Start the development server:

   ```bash
   npm run dev
   ```

### Production Deployment

- **Database**: When deploying to platforms like Vercel, replace the SQLite provider in `prisma/schema.prisma` with **PostgreSQL** or **Turso** to ensure data persistence across serverless restarts.
- **Environment Variables**:
  - `DATABASE_URL`: Connection string for your production database.
  - `ADMIN_PASSWORD`: Set this for the `/admin` login gate (default in code is `admin123`).

## License

MIT
