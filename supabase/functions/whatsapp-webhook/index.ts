
import { createClient } from "supabase";

const corsHeaders = {
    "Access-Control-Allow-Origin": "*",
    "Access-Control-Allow-Headers": "authorization, x-client-info, apikey, content-type",
};

Deno.serve(async (req) => {
    const { method } = req;

    // Handle Meta Webhook Verification (GET)
    if (method === "GET") {
        const url = new URL(req.url);
        const mode = url.searchParams.get("hub.mode");
        const token = url.searchParams.get("hub.verify_token");
        const challenge = url.searchParams.get("hub.challenge");

        if (mode === "subscribe" && token === Deno.env.get("WHATSAPP_VERIFY_TOKEN")) {
            return new Response(challenge, { status: 200 });
        }
        return new Response("Verification failed", { status: 403 });
    }

    // Handle Meta Webhook Messages (POST)
    if (method === "POST") {
        try {
            const payload = await req.json();

            // Safety: Ignore non-message updates
            const entry = payload.entry?.[0];
            const change = entry?.changes?.[0];
            const value = change?.value;
            const message = value?.messages?.[0];

            if (!message || message.type !== "text") {
                return new Response("Not a text message", { status: 200 });
            }

            const body = message.text.body;
            const whatsappMessageId = message.id;
            const from = message.from;

            // Group Restriction: Optional WHATSAPP_GROUP_ID filter
            const allowedGroupId = Deno.env.get("WHATSAPP_GROUP_ID");
            if (allowedGroupId && from !== allowedGroupId) {
                console.log(`Ignoring message from unauthorized source: ${from}`);
                return new Response("Unauthorized source", { status: 200 });
            }

            // Trigger Logic: !INFO or #INFO
            if (!body.toUpperCase().startsWith("!INFO") && !body.toUpperCase().startsWith("#INFO")) {
                return new Response("No trigger found", { status: 200 });
            }

            // Extraction Logic (Regex) - Supporting user's specific fields & case-insensitivity
            const extract = (regex: RegExp) => body.match(regex)?.[1]?.trim() || null;

            const receiverName = extract(/Receivers?\s*Name:\s*(.*)/i);
            const receiverAddress = extract(/Receivers?\s*Address:\s*(.*)/i);
            const receiverPhone = extract(/Receivers?\s*Phone:\s*(.*)/i);
            const receiverCountry = extract(/Rec[ei]{2}vers?\s*Country:\s*(.*)/i) || extract(/Destination:\s*(.*)/i);
            const senderName = extract(/Senders?\s*Name:\s*(.*)/i) || extract(/Sender:\s*(.*)/i);
            const senderCountry = extract(/Senders?\s*Country:\s*(.*)/i) || extract(/Origin:\s*(.*)/i);

            // Validation Logic: Ensure all required fields are present
            const requiredFields = [
                { key: 'receiverName', label: 'Receivers Name', value: receiverName },
                { key: 'receiverAddress', label: 'Receivers Address', value: receiverAddress },
                { key: 'receiverPhone', label: 'Receivers Phone', value: receiverPhone },
                { key: 'receiverCountry', label: 'Recievers Country', value: receiverCountry },
                { key: 'senderName', label: 'Senders Name', value: senderName },
                { key: 'senderCountry', label: 'Senders Country', value: senderCountry },
            ];

            const missingFields = requiredFields
                .filter(f => !f.value)
                .map(f => `• ${f.label}`);

            if (missingFields.length > 0) {
                console.log(`Validation failed: Missing ${missingFields.length} fields`);

                await fetch(`https://graph.facebook.com/v17.0/${Deno.env.get("WHATSAPP_PHONE_NUMBER_ID")}/messages`, {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json",
                        Authorization: `Bearer ${Deno.env.get("WHATSAPP_TOKEN")}`,
                    },
                    body: JSON.stringify({
                        messaging_product: "whatsapp",
                        recipient_type: "individual",
                        to: from,
                        context: { message_id: whatsappMessageId },
                        type: "text",
                        text: {
                            body: `⚠️ *Uplink Interrupted*\n\nManifest incomplete. Please provide the following missing fields:\n\n${missingFields.join('\n')}\n\nMaintenance Protocol: WAITING`
                        },
                    }),
                });

                return new Response(JSON.stringify({ error: "Missing required fields", missingFields }), {
                    headers: { ...corsHeaders, "Content-Type": "application/json" },
                    status: 200,
                });
            }

            // Initialize Supabase Client
            const supabase = createClient(
                Deno.env.get("SUPABASE_URL")!,
                Deno.env.get("SUPABASE_SERVICE_ROLE_KEY")!
            );

            // Duplicate Detection: Check Shipment table for existing matching manifests
            const { data: existingShipment, error: searchError } = await supabase
                .from("Shipment")
                .select("trackingNumber")
                .eq("receiverPhone", receiverPhone)
                .eq("receiverName", receiverName)
                .eq("senderName", senderName)
                .eq("receiverCountry", receiverCountry)
                .maybeSingle();

            if (searchError) {
                console.error("Duplicate check failed:", searchError);
            }

            if (existingShipment) {
                console.log(`Duplicate found: ${existingShipment.trackingNumber}`);

                await fetch(`https://graph.facebook.com/v17.0/${Deno.env.get("WHATSAPP_PHONE_NUMBER_ID")}/messages`, {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json",
                        Authorization: `Bearer ${Deno.env.get("WHATSAPP_TOKEN")}`,
                    },
                    body: JSON.stringify({
                        messaging_product: "whatsapp",
                        recipient_type: "individual",
                        to: from,
                        context: { message_id: whatsappMessageId },
                        type: "text",
                        text: {
                            body: `⚠️ *Information matches an existing manifest.*\n\nOur system indicates this shipment is already being processed. Tracking ID: *${existingShipment.trackingNumber}*\n\nMaintenance Protocol: ACTIVE`
                        },
                    }),
                });

                return new Response(JSON.stringify({ error: "Duplicate manifest", trackingId: existingShipment.trackingNumber }), {
                    headers: { ...corsHeaders, "Content-Type": "application/json" },
                    status: 200,
                });
            }

            // Generate Orbital Tracking ID
            const trackingId = `AWB-${crypto.randomUUID()}`;

            // Persist to DB (Shipment table)
            const { error: dbError } = await supabase
                .from("Shipment")
                .insert({
                    trackingNumber: trackingId,
                    status: 'PENDING',
                    senderName,
                    senderCountry,
                    receiverName,
                    receiverPhone,
                    receiverAddress,
                    receiverCountry,
                    whatsappMessageId,
                    whatsappFrom: from,
                });

            if (dbError) throw dbError;

            // Resend Notification removed as per user request


            // Reply to WhatsApp Group (Quoting original)
            await fetch(`https://graph.facebook.com/v17.0/${Deno.env.get("WHATSAPP_PHONE_NUMBER_ID")}/messages`, {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    Authorization: `Bearer ${Deno.env.get("WHATSAPP_TOKEN")}`,
                },
                body: JSON.stringify({
                    messaging_product: "whatsapp",
                    recipient_type: "individual",
                    to: from,
                    context: { message_id: whatsappMessageId },
                    type: "text",
                    text: {
                        body: `🛸 *Uplink Established*\n\nManifest Processed Successfully.\nYour Tracking ID is: *${trackingId}*\n\nStatus: [PENDING]\nAuto-Transition: 1 HOUR`
                    },
                }),
            });

            return new Response(JSON.stringify({ success: true, trackingId }), {
                headers: { ...corsHeaders, "Content-Type": "application/json" },
                status: 200,
            });

        } catch (err) {
            console.error(err);
            return new Response(JSON.stringify({ error: err.message }), {
                headers: { ...corsHeaders, "Content-Type": "application/json" },
                status: 500,
            });
        }
    }

    return new Response("Method not allowed", { status: 405 });
});
