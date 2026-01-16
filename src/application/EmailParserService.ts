import type { CreateShipmentDto } from '../domain/Shipment';

export class EmailParserService {
    /**
     * Parses the raw email body to extract shipment details.
     * Expected format:
     * Receiver: <Name>
     * Address: <Address>
     * Country: <Country>
     * Phone: <Phone>
     * Sender: <Sender>
     */
    static parse(emailBody: string): CreateShipmentDto {
        const lines = emailBody.split('\n');
        const data: Partial<CreateShipmentDto> = {};

        lines.forEach(line => {
            const lowerLine = line.toLowerCase();
            if (lowerLine.includes('receiver:')) {
                data.receiverName = line.split(':')[1].trim();
            } else if (lowerLine.includes('address:')) {
                data.receiverAddress = line.split(':')[1].trim();
            } else if (lowerLine.includes('country:')) {
                data.receiverCountry = line.split(':')[1].trim();
            } else if (lowerLine.includes('phone:')) {
                data.receiverPhone = line.split(':')[1].trim();
            } else if (lowerLine.includes('sender:')) {
                data.senderName = line.split(':')[1].trim();
            }
        });

        // Validate critical fields
        if (!data.receiverName || !data.receiverAddress || !data.receiverCountry || !data.receiverPhone || !data.senderName) {
            throw new Error("Missing critical shipment information. Please ensure Receiver, Address, Country, Phone, and Sender are present.");
        }

        return data as CreateShipmentDto;
    }
}
