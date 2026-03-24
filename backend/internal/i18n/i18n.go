package i18n

import (
	"fmt"
	"strings"
)

type Language string

const (
	EN Language = "en"
	PT Language = "pt"
	ES Language = "es"
	DE Language = "de"
)

var Translations = map[Language]map[string]string{
	EN: {
		"receipt_receiver":    "RECEIVER",
		"receipt_sender":      "SENDER",
		"receipt_destination": "DESTINATION",
		"receipt_origin":      "ORIGIN",
		"receipt_email":       "EMAIL",
		"receipt_content":     "CONTENT",
		"receipt_weight":      "WEIGHT",
		"receipt_address":     "DELIVERY ADDRESS",
		"receipt_phone":       "CONTACT PHONE",
		"receipt_service":     "SERVICE MODE",
		"receipt_payment":     "PAYMENT METHOD",
		"receipt_dep_date":    "DEPARTURE DATE",
		"receipt_arr_date":    "ARRIVAL DATE",

		// Bot Command Messages
		"ERR_ACCESS_DENIED":      "🚫 *ACCESS DENIED*\n\n_This command is restricted to the bot owner or group admins._\n\n💡 You can use `!info [ID]` to track packages.",
		"ERR_OWNER_ONLY":        "🚫 *OWNER ACCESS ONLY*\n\n_This command is restricted to the bot owner only._",
		"ERR_NOT_FOUND":         "❌ *NOT FOUND*\n\n_Could not find any shipment with that ID._",
		"ERR_DB_ERROR":          "❌ *DATABASE ERROR*\n\n_Lookup failed. Please try again later._",
		"ERR_SYSTEM_ERROR":      "❌ *SYSTEM ERROR*\n\n_Something went wrong. Please try again later._",
		"ERR_INCORRECT_USAGE":  "⚠️ *INCORRECT USAGE*\n\n_Please check the documentation or use `!help` for guidance._",
		"ERR_CONTEXT_ERROR":    "⚠️ *CONTEXT ERROR*\n\n_I couldn't find your last shipment. Please provide the tracking ID (e.g., !edit ABC-1234 ...)._",
		"MSG_LANG_UPDATED":     "🌐 *LANGUAGE UPDATED*\n\nYour language is now set to *%s*.",
		"MSG_BROADCAST_START":  "📣 *BROADCAST INITIATED*\n\nSending to *%d* authorized groups in the background.",
		"MSG_DELETE_SUCCESS":   "🗑️ *SHIPMENT DELETED*\n\nThe shipment *%s* has been permanently removed.",
		"MSG_EDIT_SUCCESS":     "✅ *INFORMATION UPDATED*\n\n🆔 *%s*\n\n📝 *FIELDS MODIFIED:*\n• %s\n\n━━━━━━━━━━━━━━━━━━━━━━━\n_Generating your updated receipt..._",
		"MSG_STATS_HEADER":     "📊 *%s VITAL STATS*",
		"MSG_STATUS_DASHBOARD": "🖥️ *SYSTEM DASHBOARD*",
	},
	PT: {
		"receipt_receiver":    "DESTINATÁRIO",
		"receipt_sender":      "REMETENTE",
		"receipt_destination": "DESTINO",
		"receipt_origin":      "ORIGEM",
		"receipt_email":       "E-MAIL",
		"receipt_content":     "CONTEÚDO",
		"receipt_weight":      "PESO",
		"receipt_address":     "ENDEREÇO DE ENTREGA",
		"receipt_phone":       "TELEFONE DE CONTATO",
		"receipt_service":     "MODO DE SERVIÇO",
		"receipt_payment":     "MÉTODO DE PAGAMENTO",
		"receipt_dep_date":    "DATA DE PARTIDA",
		"receipt_arr_date":    "DATA DE CHEGADA",

		// Bot Command Messages
		"ERR_ACCESS_DENIED":      "🚫 *ACESSO NEGADO*\n\n_Este comando é restrito ao proprietário do bot ou administradores do grupo._\n\n💡 Você pode usar `!info [ID]` para rastrear pacotes.",
		"ERR_OWNER_ONLY":        "🚫 *APENAS PROPRIETÁRIO*\n\n_Este comando é restrito apenas ao proprietário do bot._",
		"ERR_NOT_FOUND":         "❌ *NÃO ENCONTRADO*\n\n_Não foi possível encontrar nenhum envio com esse ID._",
		"ERR_DB_ERROR":          "❌ *ERRO DE BANCO DE DADOS*\n\n_A consulta falhou. Por favor, tente novamente mais tarde._",
		"ERR_SYSTEM_ERROR":      "❌ *ERRO DE SISTEMA*\n\n_Algo deu errado. Por favor, tente novamente mais tarde._",
		"ERR_INCORRECT_USAGE":  "⚠️ *USO INCORRETO*\n\n_Por favor, verifique a documentação ou use `!help` para orientação._",
		"ERR_CONTEXT_ERROR":    "⚠️ *ERRO DE CONTEXTO*\n\n_Não consegui encontrar seu último envio. Por favor, forneça o ID de rastreamento (ex: !edit ABC-1234 ...)._",
		"MSG_LANG_UPDATED":     "🌐 *IDIOMA ATUALIZADO*\n\nSeu idioma agora está definido como *%s*.",
		"MSG_BROADCAST_START":  "📣 *TRANSMISSÃO INICIADA*\n\nEnviando para *%d* grupos autorizados em segundo plano.",
		"MSG_DELETE_SUCCESS":   "🗑️ *ENVIO EXCLUÍDO*\n\nO envio *%s* foi removido permanentemente.",
		"MSG_EDIT_SUCCESS":     "✅ *INFORMAÇÕES ATUALIZADAS*\n\n🆔 *%s*\n\n📝 *CAMPOS MODIFICADOS:*\n• %s\n\n━━━━━━━━━━━━━━━━━━━━━━━\n_Gerando seu recibo atualizado..._",
		"MSG_STATS_HEADER":     "📊 *ESTATÍSTICAS VITAIS DA %s*",
		"MSG_STATUS_DASHBOARD": "🖥️ *PAINEL DO SISTEMA*",
	},
	ES: {
		"receipt_receiver":    "DESTINATARIO",
		"receipt_sender":      "REMITENTE",
		"receipt_destination": "DESTINO",
		"receipt_origin":      "ORIGEN",
		"receipt_email":       "CORREO",
		"receipt_content":     "CONTENIDO",
		"receipt_weight":      "PESO",
		"receipt_address":     "DIRECCIÓN DE ENTREGA",
		"receipt_phone":       "TELÉFONO DE CONTACTO",
		"receipt_service":     "MODO DE SERVICIO",
		"receipt_payment":     "MÉTODO DE PAGO",
		"receipt_dep_date":    "FECHA DE SALIDA",
		"receipt_arr_date":    "FECHA DE LLEGADA",

		// Bot Command Messages
		"ERR_ACCESS_DENIED":      "🚫 *ACCESO DENEGADO*\n\n_Este comando está restringido al propietario del bot o a los administradores del grupo._\n\n💡 Puedes usar `!info [ID]` para rastrear paquetes.",
		"ERR_OWNER_ONLY":        "🚫 *SOLO PROPIETARIO*\n\n_Este comando está restringido solo al propietario del bot._",
		"ERR_NOT_FOUND":         "❌ *NO ENCONTRADO*\n\n_No se pudo encontrar ningún envío con ese ID._",
		"ERR_DB_ERROR":          "❌ *ERROR DE BASE DE DATOS*\n\n_La búsqueda falló. Por favor, inténtelo de nuevo más tarde._",
		"ERR_SYSTEM_ERROR":      "❌ *ERROR DE SISTEMA*\n\n_Algo salió mal. Por favor, inténtelo de nuevo más tarde._",
		"ERR_INCORRECT_USAGE":  "⚠️ *USO INCORRETO*\n\n_Por favor, consulte la documentación o use `!help` para obtener orientación._",
		"ERR_CONTEXT_ERROR":    "⚠️ *ERROR DE CONTEXTO*\n\n_No pude encontrar su último envío. Por favor, proporcione el ID de seguimiento (ej: !edit ABC-1234 ...)._",
		"MSG_LANG_UPDATED":     "🌐 *IDIOMA ACTUALIZADO*\n\nSu idioma ahora está configurado en *%s*.",
		"MSG_BROADCAST_START":  "📣 *TRANSMISIÓN INICIADA*\n\nEnviando a *%d* grupos autorizados en segundo plano.",
		"MSG_DELETE_SUCCESS":   "🗑️ *ENVÍO ELIMINADO*\n\nEl envío *%s* ha sido eliminado permanentemente.",
		"MSG_EDIT_SUCCESS":     "✅ *INFORMACIÓN ACTUALIZADA*\n\n🆔 *%s*\n\n📝 *INFORMACIÓN MODIFICADA:*\n• %s\n\n━━━━━━━━━━━━━━━━━━━━━━━\n_Generando su recibo actualizado..._",
		"MSG_STATS_HEADER":     "📊 *ESTADÍSTICAS VITALES DE %s*",
		"MSG_STATUS_DASHBOARD": "🖥️ *PANEL DEL SISTEMA*",
	},
	DE: {
		"receipt_receiver":    "EMPFÄNGER",
		"receipt_sender":      "ABSENDER",
		"receipt_destination": "ZIELORT",
		"receipt_origin":      "HERKUNFT",
		"receipt_email":       "E-MAIL",
		"receipt_content":     "INHALT",
		"receipt_weight":      "GEWICHT",
		"receipt_address":     "LIEFERADRESSE",
		"receipt_phone":       "KONTAKTTELEFON",
		"receipt_service":     "SERVICEART",
		"receipt_payment":     "ZAHLUNGSART",
		"receipt_dep_date":    "ABFAHRTSDATUM",
		"receipt_arr_date":    "ANKUNFTSDATUM",

		// Bot Command Messages
		"ERR_ACCESS_DENIED":      "🚫 *ZUGRIFF VERWEIGERT*\n\n_Dieser Befehl ist dem Bot-Besitzer oder Gruppenadministratoren vorbehalten._\n\n💡 Sie können `!info [ID]` verwenden, um Pakete zu verfolgen.",
		"ERR_OWNER_ONLY":        "🚫 *NUR BESITZER*\n\n_Dieser Befehl ist nur dem Bot-Besitzer vorbehalten._",
		"ERR_NOT_FOUND":         "❌ *NICHT GEFUNDEN*\n\n_Es konnte keine Sendung mit dieser ID gefunden werden._",
		"ERR_DB_ERROR":          "❌ *DATENBANKFEHLER*\n\n_Suche fehlgeschlagen. Bitte versuchen Sie es später erneut._",
		"ERR_SYSTEM_ERROR":      "❌ *SYSTEMFEHLER*\n\n_Etwas ist schief gelaufen. Bitte versuchen Sie es später erneut._",
		"ERR_INCORRECT_USAGE":  "⚠️ *FALSCHE VERWENDUNG*\n\n_Bitte überprüfen Sie die Dokumentation oder verwenden Sie `!help` zur Orientierung._",
		"ERR_CONTEXT_ERROR":    "⚠️ *KONTEXTFEHLER*\n\n_Ich konnte Ihre letzte Sendung nicht finden. Bitte geben Sie die Tracking-ID an (z. B. !edit ABC-1234 ...)._",
		"MSG_LANG_UPDATED":     "🌐 *SPRACHE AKTUALISIERT*\n\nIhre Sprache ist jetzt auf *%s* eingestellt.",
		"MSG_BROADCAST_START":  "📣 *BROADCAST GESTARTET*\n\nSenden an *%d* autorisierte Gruppen im Hintergrund.",
		"MSG_DELETE_SUCCESS":   "🗑️ *SENDUNG GELÖSCHT*\n\nDie Sendung *%s* wurde dauerhaft entfernt.",
		"MSG_EDIT_SUCCESS":     "✅ *INFORMATIONEN AKTUALISIERT*\n\n🆔 *%s*\n\n📝 *GEÄNDERTE FELDER:*\n• %s\n\n━━━━━━━━━━━━━━━━━━━━━━━\n_Ihr aktualisierter Beleg wird generiert..._",
		"MSG_STATS_HEADER":     "📊 *WICHTIGE STATISTIKEN VON %s*",
		"MSG_STATUS_DASHBOARD": "🖥️ *SYSTEM-DASHBOARD*",
	},
}

func T(lang Language, key string, args ...interface{}) string {
	l := Language(strings.ToLower(string(lang)))
	if _, ok := Translations[l]; !ok {
		l = EN
	}
	val, ok := Translations[l][key]
	if !ok {
		// Fallback to EN for specific key
		val = Translations[EN][key]
		if val == "" {
			return fmt.Sprintf("!! %s !!", key)
		}
	}

	if len(args) > 0 {
		return fmt.Sprintf(val, args...)
	}
	return val
}

func GetDateFormat(lang Language) string {
	l := Language(strings.ToLower(string(lang)))
	switch l {
	case DE:
		return "02.01.2006"
	case EN:
		return "02/01/2006"
	case ES, PT:
		return "02/01/2006"
	default:
		return "02/01/2006"
	}
}
