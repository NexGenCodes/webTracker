'use server'

import { prisma } from '@/lib/prisma'
import { CreateShipmentDto } from '@/lib/email-parser'
import { revalidatePath } from 'next/cache'
import { ShipmentData } from '@/types/shipment'
import crypto from 'crypto'

export async function createShipment(data: CreateShipmentDto) {
    try {
        const hashInput = `${data.receiverName}${data.receiverPhone}${data.receiverCountry}${data.senderName}`;
        const hash = crypto.createHash('shake256', { outputLength: 5 })
            .update(hashInput)
            .digest('hex');

        const trackingNumber = `AWB-${hash}`;

        const shipment = await prisma.shipment.create({
            data: {
                trackingNumber,
                status: 'PENDING',
                senderName: data.senderName,
                receiverName: data.receiverName,
                receiverAddress: data.receiverAddress,
                receiverCountry: data.receiverCountry,
                receiverPhone: data.receiverPhone,
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

export async function getTracking(trackingNumber: string): Promise<ShipmentData | null> {
    if (!trackingNumber) return null

    const shipment = await prisma.shipment.findUnique({
        where: { trackingNumber },
        include: { events: { orderBy: { timestamp: 'desc' } } }
    })

    if (!shipment) return null

    if (shipment.isArchived) {
        return {
            trackingNumber: shipment.trackingNumber,
            status: 'DELIVERED',
            isArchived: true,
            senderName: null,
            receiverName: null,
            receiverAddress: null,
            receiverCountry: null,
            receiverPhone: null,
            id: shipment.id,
            events: []
        } as ShipmentData;
    }

    return shipment as unknown as ShipmentData;
}

export async function markAsDelivered(trackingNumber: string) {
    const shipment = await prisma.shipment.findUnique({ where: { trackingNumber } })
    if (!shipment) return { success: false, error: 'Not found' }

    await prisma.$transaction([
        prisma.event.create({
            data: {
                shipmentId: shipment.id,
                status: 'DELIVERED',
                location: 'Destination',
                notes: 'Delivered to recipient'
            }
        }),
        prisma.shipment.update({
            where: { id: shipment.id },
            data: {
                status: 'DELIVERED',
                isArchived: true,
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
