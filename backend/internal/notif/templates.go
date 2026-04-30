package notif

import (
	"fmt"
	"time"
)

// ---------------------------------------------------------------------------
// Template helpers — every public func returns an Email value.
// The caller can Send() or SendAsync() it via a Mailer.
// ---------------------------------------------------------------------------

// OTPEmail builds the 6-digit verification email.
func OTPEmail(to, otp string) Email {
	companyName := "CargoHive"
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #f4f4f4; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: auto; background: white; padding: 40px; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
        <h1 style="color: #2563eb; font-size: 24px; margin-bottom: 10px;">%s</h1>
        <h2 style="color: #333; font-size: 18px; margin-bottom: 25px;">Verify your email address</h2>
        <p style="color: #555; line-height: 1.6;">Please use the following verification code to complete your registration. This code is valid for <strong>10 minutes</strong>.</p>
        <div style="text-align: center; margin: 35px 0; padding: 25px; background: #f8f9fa; border-radius: 8px;">
            <p style="margin: 0 0 10px 0; font-size: 14px; text-transform: uppercase; letter-spacing: 1px; color: #666;">Your Verification Code</p>
            <div style="font-size: 42px; font-weight: bold; letter-spacing: 8px; color: #000; font-family: 'Courier New', Courier, monospace;">%s</div>
        </div>
        <p style="margin-top: 30px; border-top: 1px solid #eee; padding-top: 20px; color: #888; font-size: 12px;">
            If you didn't request this code, you can safely ignore this email.
            <br>&copy; %d %s. All rights reserved.
        </p>
    </div>
</body>
</html>`, companyName, otp, time.Now().Year(), companyName)

	return Email{
		To:       to,
		Subject:  "Your Verification Code",
		HTMLBody: html,
		FromName: companyName,
	}
}

// SetupLinkEmail builds the magic-link onboarding email.
func SetupLinkEmail(to, companyName, frontendURL, token string) Email {
	if companyName == "" {
		companyName = "AIRWAYBILL"
	}
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	link := fmt.Sprintf("%s/setup/%s", frontendURL, token)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #f4f4f4; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: auto; background: white; padding: 40px; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
        <h1 style="color: #2563eb; font-size: 24px; margin-bottom: 10px;">%s</h1>
        <h2 style="color: #333; font-size: 18px; margin-bottom: 25px;">WhatsApp Bot Setup</h2>
        <p style="color: #555; line-height: 1.6;">Welcome! Your account has been created. Click the button below to complete your setup by linking your WhatsApp Business number.</p>
        <div style="text-align: center; margin: 35px 0;">
            <a href="%s" style="display: inline-block; padding: 16px 40px; background: linear-gradient(135deg, #2563eb, #1e40af); color: white; text-decoration: none; border-radius: 12px; font-weight: 700; font-size: 16px; letter-spacing: 0.5px;">
                Complete Setup →
            </a>
        </div>
        <div style="background-color: #f8f9fa; border-left: 4px solid #2563eb; padding: 15px; margin: 25px 0;">
            <p style="margin: 0; color: #333; font-size: 13px;"><strong>Direct Link:</strong></p>
            <p style="margin: 5px 0 0 0; color: #2563eb; word-break: break-all; font-size: 13px;">%s</p>
        </div>
        <p style="color: #dc3545; font-size: 14px; font-weight: bold; text-align: center;">🔒 This link is unique to your company. Do not share it.</p>
        <p style="margin-top: 30px; border-top: 1px solid #eee; padding-top: 20px; color: #888; font-size: 12px;">
            If you did not request this, please ignore this email.
            <br>&copy; 2026 %s. All rights reserved.
        </p>
    </div>
</body>
</html>`, companyName, link, link, companyName)

	return Email{
		To:       to,
		Subject:  fmt.Sprintf("[%s] Complete Your WhatsApp Bot Setup", companyName),
		HTMLBody: html,
		FromName: companyName,
	}
}

// PairingCodeEmail builds the WhatsApp pairing-code email.
func PairingCodeEmail(to, companyName, phone, code string) Email {
	companyName = "CargoHive"
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: auto; background: white; padding: 40px; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
        <h1 style="color: #007bff; font-size: 24px; margin-bottom: 10px;">%s</h1>
        <h2 style="color: #333; font-size: 18px; margin-bottom: 25px;">WhatsApp Integration Setup</h2>
        <p style="color: #555; line-height: 1.6;">You have requested a pairing code to link your WhatsApp account to the tracking bot.</p>
        <div style="background-color: #f8f9fa; border-left: 4px solid #007bff; padding: 15px; margin: 25px 0;">
            <p style="margin: 0; color: #333;"><strong>Target Phone:</strong> %s</p>
        </div>
        <div style="text-align: center; margin: 35px 0; padding: 25px; background: #e9ecef; border-radius: 8px;">
            <p style="margin: 0 0 10px 0; font-size: 14px; text-transform: uppercase; letter-spacing: 1px; color: #666;">Your Pairing Code</p>
            <div style="font-size: 42px; font-weight: bold; letter-spacing: 8px; color: #000; font-family: 'Courier New', Courier, monospace;">%s</div>
        </div>
        <p style="color: #dc3545; font-size: 14px; font-weight: bold; text-align: center;">⏱️ THIS CODE EXPIRES IN 2 MINUTES</p>
        <p style="margin-top: 30px; border-top: 1px solid #eee; padding-top: 20px; color: #888; font-size: 12px;">
            If you did not request this code, please ignore this email or contact security.
            <br>&copy; 2026 %s. All rights reserved.
        </p>
    </div>
</body>
</html>`, companyName, phone, code, companyName)

	return Email{
		To:       to,
		Subject:  fmt.Sprintf("[%s] WhatsApp Pairing Code", companyName),
		HTMLBody: html,
		FromName: companyName,
	}
}

// DeliveryEmail builds the shipment-arrival notification email.
func DeliveryEmail(to, recipientName, trackingID, companyName, arrivalDate string) Email {
	if companyName == "" {
		companyName = "AIRWAYBILL"
	}
	if recipientName == "" {
		recipientName = "Customer"
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: auto; background: white; padding: 40px; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
        <h1 style="color: #28a745; font-size: 24px; margin-bottom: 10px;">%s</h1>
        <h2 style="color: #333; font-size: 18px; margin-bottom: 25px;">Notice of Arrival</h2>
        <p style="color: #555; line-height: 1.6;">Hello <strong>%s</strong>,</p>
        <p style="color: #555; line-height: 1.6;">This is an official notification to inform you that your package (Tracking ID: <strong>%s</strong>) has successfully arrived in your country.</p>
        <p style="color: #555; line-height: 1.6;">It is currently securely held at our local depot. One of our regional dispatchers will be contacting you shortly to coordinate the final delivery details to your address.</p>
        <div style="background-color: #f8f9fa; border-left: 4px solid #28a745; padding: 15px; margin: 25px 0;">
            <p style="margin: 0; color: #333;"><strong>Tracking ID:</strong> %s</p>
            <p style="margin: 5px 0 0 0; color: #333;"><strong>Status:</strong> ARRIVED AT DESTINATION</p>
            <p style="margin: 5px 0 0 0; color: #333;"><strong>Arrival Date:</strong> %s</p>
        </div>
        <p style="color: #555; line-height: 1.6;">Thank you for choosing %s. We appreciate your patience during this final transit phase.</p>
        <p style="margin-top: 30px; border-top: 1px solid #eee; padding-top: 20px; color: #888; font-size: 12px;">
            This is an automated message. Please await further contact from our local agent.
            <br>&copy; 2026 %s. All rights reserved.
        </p>
    </div>
</body>
</html>`, companyName, recipientName, trackingID, trackingID, arrivalDate, companyName, companyName)

	return Email{
		To:       to,
		Subject:  fmt.Sprintf("[%s] Package Arrival Notification - %s", companyName, trackingID),
		HTMLBody: html,
		FromName: companyName,
	}
}

// PasswordResetEmail builds the password reset OTP email.
func PasswordResetEmail(to, otp string) Email {
	companyName := "CargoHive"
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; margin: 0; padding: 20px;">
    <div style="max-width: 600px; margin: auto; background: white; padding: 40px; border-radius: 12px; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
        <h1 style="color: #dc3545; font-size: 24px; margin-bottom: 10px;">%s</h1>
        <h2 style="color: #333; font-size: 18px; margin-bottom: 25px;">Password Reset Request</h2>
        <p style="color: #555; line-height: 1.6;">We received a request to reset your password. Use the following code to proceed. This code is valid for <strong>15 minutes</strong>.</p>
        <div style="text-align: center; margin: 35px 0; padding: 25px; background: #fff5f5; border-radius: 8px; border: 1px solid #feb2b2;">
            <p style="margin: 0 0 10px 0; font-size: 14px; text-transform: uppercase; letter-spacing: 1px; color: #c53030;">Your Reset Code</p>
            <div style="font-size: 42px; font-weight: bold; letter-spacing: 8px; color: #c53030; font-family: 'Courier New', Courier, monospace;">%s</div>
        </div>
        <p style="color: #555; font-size: 14px; text-align: center;">If you did not request a password reset, you can safely ignore this email.</p>
        <p style="margin-top: 30px; border-top: 1px solid #eee; padding-top: 20px; color: #888; font-size: 12px;">
            &copy; 2026 %s. All rights reserved.
        </p>
    </div>
</body>
</html>`, companyName, otp, companyName)

	return Email{
		To:       to,
		Subject:  "Reset Your Password",
		HTMLBody: html,
		FromName: companyName,
	}
}

