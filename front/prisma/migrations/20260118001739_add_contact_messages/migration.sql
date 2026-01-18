/*
  Warnings:

  - A unique constraint covering the columns `[whatsappMessageId]` on the table `Shipment` will be added. If there are existing duplicate values, this will fail.

*/
-- AlterTable
ALTER TABLE "Shipment" ADD COLUMN     "lastNotifiedAt" TIMESTAMP(3),
ADD COLUMN     "lastTransitionAt" TIMESTAMP(3),
ADD COLUMN     "receiverID" TEXT,
ADD COLUMN     "senderCountry" TEXT,
ADD COLUMN     "whatsappFrom" TEXT,
ADD COLUMN     "whatsappMessageId" TEXT;

-- CreateTable
CREATE TABLE "ContactMessage" (
    "id" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "email" TEXT NOT NULL,
    "message" TEXT NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "ContactMessage_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "Shipment_whatsappMessageId_key" ON "Shipment"("whatsappMessageId");
