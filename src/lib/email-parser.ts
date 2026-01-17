export type CreateShipmentDto = {
    receiverName: string;
    receiverAddress: string;
    receiverCountry: string;
    receiverPhone: string;
    senderName: string;
    senderCountry: string;
}

export function parseEmail(emailBody: string): CreateShipmentDto {
    const data: Partial<CreateShipmentDto> = {};

    // Define regex patterns for more robust matching
    const patterns = {
        receiverName: /(?:Receiver|To|Recipient):\s*(.*)/i,
        receiverAddress: /(?:Address|Delivery Address|Location):\s*(.*)/i,
        receiverCountry: /(?:Country|Destination):\s*(.*)/i,
        receiverPhone: /(?:Phone|Contact|Tel):\s*(.*)/i,
        senderName: /(?:Sender|From|Originator):\s*(.*)/i,
        senderCountry: /(?:Sender Country|Origin Country|From Country):\s*(.*)/i,
    };

    Object.entries(patterns).forEach(([key, regex]) => {
        const match = emailBody.match(regex);
        if (match && match[1]) {
            data[key as keyof CreateShipmentDto] = match[1].trim();
        }
    });

    // Final validation
    const missing = [];
    if (!data.receiverName) missing.push('Receiver Name');
    if (!data.receiverAddress) missing.push('Address');
    if (!data.receiverCountry) missing.push('Country');
    if (!data.receiverPhone) missing.push('Phone');
    if (!data.senderName) missing.push('Sender Name');
    if (!data.senderCountry) missing.push('Sender Country');

    if (missing.length > 0) {
        throw new Error(`Missing information: ${missing.join(', ')}. Please Ensure these labels are present in the email.`);
    }

    return data as CreateShipmentDto;
}
