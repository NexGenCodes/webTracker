-- Database Schema Reference (Source of Truth for Go Bot)
-- Synchronize this with front/prisma/schema.prisma after migrations.

CREATE TABLE "Shipment" (
    "id" TEXT PRIMARY KEY,
    "trackingNumber" TEXT UNIQUE NOT NULL,
    "status" TEXT DEFAULT 'PENDING' NOT NULL,
    "senderName" TEXT,
    "senderCountry" TEXT,
    "receiverName" TEXT,
    "receiverPhone" TEXT,
    "receiverAddress" TEXT,
    "receiverCountry" TEXT,
    "whatsappFrom" TEXT,
    "createdAt" TIMESTAMPTZ DEFAULT NOW(),
    "updatedAt" TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE "Event" (
    "id" TEXT PRIMARY KEY,
    "shipmentId" TEXT REFERENCES "Shipment"("id") ON DELETE CASCADE,
    "status" TEXT NOT NULL,
    "description" TEXT,
    "location" TEXT,
    "createdAt" TIMESTAMPTZ DEFAULT NOW()
);

-- Indices for performance
CREATE INDEX idx_shipment_tracking ON "Shipment"("trackingNumber");
CREATE INDEX idx_shipment_status_created ON "Shipment"("status", "createdAt");
