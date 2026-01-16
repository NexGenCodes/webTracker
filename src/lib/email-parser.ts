export type CreateShipmentDto = {
    receiverName: string;
    receiverAddress: string;
    receiverCountry: string;
    receiverPhone: string;
    senderName: string;
}

export function parseEmail(emailBody: string): CreateShipmentDto {
    const lines = emailBody.split('\n');
    const data: Partial<CreateShipmentDto> = {};

    lines.forEach(line => {
        const lowerLine = line.toLowerCase();
        // Simple case-insensitive check
        if (lowerLine.includes('receiver:')) {
            data.receiverName = line.split(':')[1]?.trim();
        } else if (lowerLine.includes('address:')) {
            data.receiverAddress = line.split(':')[1]?.trim();
        } else if (lowerLine.includes('country:')) {
            data.receiverCountry = line.split(':')[1]?.trim();
        } else if (lowerLine.includes('phone:')) {
            data.receiverPhone = line.split(':')[1]?.trim();
        } else if (lowerLine.includes('sender:')) {
            data.senderName = line.split(':')[1]?.trim();
        }
    });

    if (!data.receiverName || !data.receiverAddress || !data.receiverCountry || !data.receiverPhone || !data.senderName) {
        throw new Error("Missing critical information. Ensure Receiver, Address, Country, Phone, and Sender are present.");
    }

    return data as CreateShipmentDto;
}
