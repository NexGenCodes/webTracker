'use server'

import { prisma } from '@/lib/prisma'
import { CreateShipmentDto } from '@/lib/email-parser'
import { revalidatePath } from 'next/cache'

export async function createShipment(data: CreateShipmentDto) {
    try {
        // Generate a simple tracking ID (e.g., TRK-TIMESTAMP-RANDOM) 
        // or use CUID. User requested "tracking number", let's make it look nice.
        const trackingNumber = `TRK-${Date.now().toString().slice(-6)}-${Math.floor(Math.random() * 1000)}`

        const shipment = await prisma.shipment.create({
            data: {
                trackingNumber,
                status: 'PENDING',
                senderName: data.senderName,
                receiverName: data.receiverName,
                receiverAddress: data.receiverAddress,
                receiverCountry: data.receiverCountry,
                receiverPhone: data.receiverPhone,
                // Create initial event
                events: {
                    create: {
                        status: 'PENDING',
                        location: data.senderName.includes('Global') ? 'Warehouse' : 'Origin',
                        notes: 'Shipment created'
                    }
                }
            }
        })

        return { success: true, trackingNumber: shipment.trackingNumber }
    } catch (error) {
        console.error('Failed to create shipment:', error)
        return { success: false, error: 'Database error' }
    }
}

export async function getTracking(trackingNumber: string) {
    if (!trackingNumber) return null

    const shipment = await prisma.shipment.findUnique({
        where: { trackingNumber },
        include: { events: { orderBy: { timestamp: 'desc' } } }
    })

    if (!shipment) return null

    // ARCHIVE LOGIC:
    // If archived, we return a sanitized object
    if (shipment.isArchived) {
        return {
            trackingNumber: shipment.trackingNumber,
            status: 'DELIVERED',
            isArchived: true,
            // Hide all PII
            senderName: null,
            receiverName: null,
            receiverAddress: null,
            receiverCountry: null, // Maybe keep country? User said "deleted". Safer to hide all.
            events: [] // Hide history? Or keep simple "Delivered"?
            // User requirement: "data is deleted... user should get delivered"
        }
    }

    return shipment
}

// Additional helper to simulate "Mark as Delivered" + Archive
export async function markAsDelivered(trackingNumber: string) {
    const shipment = await prisma.shipment.findUnique({ where: { trackingNumber } })
    if (!shipment) return { success: false, error: 'Not found' }

    await prisma.$transaction([
        // Create Delivered Event
        prisma.event.create({
            data: {
                shipmentId: shipment.id,
                status: 'DELIVERED',
                location: 'Destination',
                notes: 'Delivered to recipient'
            }
        }),
        // Update Shipment and SCRUB PII
        prisma.shipment.update({
            where: { id: shipment.id },
            data: {
                status: 'DELIVERED',
                isArchived: true,
                // Scrub data
                senderName: null,
                receiverName: null,
                receiverEmail: null,
                receiverAddress: null,
                receiverCountry: null,
                receiverPhone: null,
            }
        })
    ])

    revalidatePath('/')
    return { success: true }
}
