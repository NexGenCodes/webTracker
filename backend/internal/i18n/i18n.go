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
		"receipt_content":     "CONTENTS",
		"receipt_weight":      "WEIGHT",
		"receipt_address":     "DELIVERY ADDRESS",
		"receipt_phone":       "CONTACT NUMBER",
		"receipt_service":     "SERVICE TYPE",
		"receipt_payment":     "PAYMENT METHOD",
		"receipt_dep_date":    "DISPATCH DATE",
		"receipt_arr_date":    "EST. ARRIVAL",

		"ERR_ACCESS_DENIED":    "🚫 *Access Denied*\n\n_This operation is restricted to system administrators._\n\n💡 _Tip: You can use `!info [ID]` to track a shipment._",
		"ERR_OWNER_ONLY":       "🔐 *System Access Restricted*\n\n_This administrative command is reserved for the system owner._",
		"ERR_NOT_FOUND":        "🔍 *Record Not Found*\n\n_We could not locate a shipment matching that ID. Please verify and try again._",
		"ERR_DB_ERROR":         "⚠️ *Service Interruption*\n\n_We are experiencing a temporary database issue. Please try again in a few moments._",
		"ERR_SYSTEM_ERROR":     "⚠️ *System Error*\n\n_An unexpected error occurred while processing your request. Our support team has been notified._",
		"ERR_INCORRECT_USAGE":  "📝 *Invalid Command Format*\n\n_Please verify the command syntax. Reply with `!help` to view the documentation._",
		"ERR_CONTEXT_ERROR":    "⏳ *Session Expired*\n\n_Could not determine the context of your request. Please specify the Tracking ID (e.g., `!edit ABC-1234 ...`)._",
		"MSG_LANG_UPDATED":     "🌐 *Preferences Updated*\n\n_Your interface language has been successfully set to *%s*._",
		"MSG_BROADCAST_START":  "📣 *Broadcast Initiated*\n\n_Your message is being securely transmitted to *%d* authorized groups._",
		"MSG_DELETE_SUCCESS":   "🗑️ *Record Deleted*\n\n_The logistics record for *%s* has been securely removed._",
		"MSG_EDIT_SUCCESS":     "✅ *Shipment Details Updated*\n\n🆔 *%s*\n\n📝 *Modified Fields:*\n• %s\n\n━━━━━━━━━━━━━━━━━━━━━━━\n_Please wait while we generate your updated digital receipt..._",
		"MSG_STATS_HEADER":     "📊 *%s System Metrics*",
		"MSG_STATUS_DASHBOARD": "🖥️ *Operations Dashboard*",
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
		"receipt_service":     "TIPO DE SERVIÇO",
		"receipt_payment":     "MÉTODO DE PAGAMENTO",
		"receipt_dep_date":    "DATA DE ENVIO",
		"receipt_arr_date":    "CHEGADA ESTIMADA",

		"ERR_ACCESS_DENIED":    "🚫 *Acesso Negado*\n\n_Esta operação é restrita aos administradores do sistema._\n\n💡 _Dica: Você pode usar `!info [ID]` para rastrear um envio._",
		"ERR_OWNER_ONLY":       "🔐 *Acesso Restrito*\n\n_Este comando administrativo é reservado ao proprietário do sistema._",
		"ERR_NOT_FOUND":        "🔍 *Registro Não Encontrado*\n\n_Não foi possível localizar um envio correspondente a este ID. Por favor, verifique e tente novamente._",
		"ERR_DB_ERROR":         "⚠️ *Interrupção de Serviço*\n\n_Estamos enfrentando um problema temporário no banco de dados. Por favor, tente novamente em alguns instantes._",
		"ERR_SYSTEM_ERROR":     "⚠️ *Erro de Sistema*\n\n_Ocorreu um erro inesperado ao processar sua solicitação. Nossa equipe de suporte foi notificada._",
		"ERR_INCORRECT_USAGE":  "📝 *Formato de Comando Inválido*\n\n_Por favor, verifique a sintaxe do comando. Responda com `!help` para visualizar a documentação._",
		"ERR_CONTEXT_ERROR":    "⏳ *Sessão Expirada*\n\n_Não foi possível determinar o contexto da sua solicitação. Por favor, especifique o ID de Rastreamento (ex: `!edit ABC-1234 ...`)._",
		"MSG_LANG_UPDATED":     "🌐 *Preferências Atualizadas*\n\n_O idioma da interface foi configurado com sucesso para *%s*._",
		"MSG_BROADCAST_START":  "📣 *Transmissão Iniciada*\n\n_Sua mensagem está sendo transmitida com segurança para *%d* grupos autorizados._",
		"MSG_DELETE_SUCCESS":   "🗑️ *Registro Excluído*\n\n_O registro logístico para *%s* foi removido com segurança._",
		"MSG_EDIT_SUCCESS":     "✅ *Detalhes do Envio Atualizados*\n\n🆔 *%s*\n\n📝 *Campos Modificados:*\n• %s\n\n━━━━━━━━━━━━━━━━━━━━━━━\n_Por favor, aguarde enquanto geramos seu recibo digital atualizado..._",
		"MSG_STATS_HEADER":     "📊 *Métricas do Sistema %s*",
		"MSG_STATUS_DASHBOARD": "🖥️ *Painel de Operações*",
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
		"receipt_service":     "TIPO DE SERVICIO",
		"receipt_payment":     "MÉTODO DE PAGO",
		"receipt_dep_date":    "FECHA DE ENVÍO",
		"receipt_arr_date":    "LLEGADA ESTIMADA",

		"ERR_ACCESS_DENIED":    "🚫 *Acceso Denegado*\n\n_Esta operación está restringida a los administradores del sistema._\n\n💡 _Consejo: Puede usar `!info [ID]` para rastrear un envío._",
		"ERR_OWNER_ONLY":       "🔐 *Acceso Restringido*\n\n_Este comando administrativo está reservado para el propietario del sistema._",
		"ERR_NOT_FOUND":        "🔍 *Registro No Encontrado*\n\n_No pudimos localizar un envío que coincida con este ID. Por favor, verifique e inténtelo de nuevo._",
		"ERR_DB_ERROR":         "⚠️ *Interrupción del Servicio*\n\n_Estamos experimentando un problema temporal en la base de datos. Por favor, inténtelo de nuevo en unos momentos._",
		"ERR_SYSTEM_ERROR":     "⚠️ *Error del Sistema*\n\n_Ocurrió un error inesperado al procesar su solicitud. Nuestro equipo de soporte ha sido notificado._",
		"ERR_INCORRECT_USAGE":  "📝 *Formato de Comando Inválido*\n\n_Por favor, verifique la sintaxis del comando. Responda con `!help` para ver la documentación._",
		"ERR_CONTEXT_ERROR":    "⏳ *Sesión Caducada*\n\n_No se pudo determinar el contexto de su solicitud. Por favor, especifique el ID de Seguimiento (ej: `!edit ABC-1234 ...`)._",
		"MSG_LANG_UPDATED":     "🌐 *Preferencias Actualizadas*\n\n_El idioma de la interfaz se ha configurado correctamente a *%s*._",
		"MSG_BROADCAST_START":  "📣 *Transmisión Iniciada*\n\n_Su mensaje está siendo transmitido de forma segura a *%d* grupos autorizados._",
		"MSG_DELETE_SUCCESS":   "🗑️ *Registro Eliminado*\n\n_El registro logístico de *%s* ha sido eliminado de forma segura._",
		"MSG_EDIT_SUCCESS":     "✅ *Detalles de Envío Actualizados*\n\n🆔 *%s*\n\n📝 *Campos Modificados:*\n• %s\n\n━━━━━━━━━━━━━━━━━━━━━━━\n_Por favor, espere mientras generamos su recibo digital actualizado..._",
		"MSG_STATS_HEADER":     "📊 *Métricas del Sistema %s*",
		"MSG_STATUS_DASHBOARD": "🖥️ *Panel de Operaciones*",
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
		"receipt_phone":       "KONTAKTNUMMER",
		"receipt_service":     "SERVICEART",
		"receipt_payment":     "ZAHLUNGSMETHODE",
		"receipt_dep_date":    "VERSANDDATUM",
		"receipt_arr_date":    "VORAUSSICHTLICHE ANKUNFT",

		"ERR_ACCESS_DENIED":    "🚫 *Zugriff Verweigert*\n\n_Dieser Vorgang ist auf Systemadministratoren beschränkt._\n\n💡 _Tipp: Sie können `!info [ID]` verwenden, um eine Sendung zu verfolgen._",
		"ERR_OWNER_ONLY":       "🔐 *Eingeschränkter Zugriff*\n\n_Dieser administrative Befehl ist dem Systembesitzer vorbehalten._",
		"ERR_NOT_FOUND":        "🔍 *Eintrag Nicht Gefunden*\n\n_Wir konnten keine Sendung mit dieser ID finden. Bitte überprüfen Sie die Eingabe._",
		"ERR_DB_ERROR":         "⚠️ *Dienstunterbrechung*\n\n_Es tritt ein vorübergehendes Datenbankproblem auf. Bitte versuchen Sie es in Kürze erneut._",
		"ERR_SYSTEM_ERROR":     "⚠️ *Systemfehler*\n\n_Bei der Verarbeitung Ihrer Anfrage ist ein unerwarteter Fehler aufgetreten. Unser Support-Team wurde benachrichtigt._",
		"ERR_INCORRECT_USAGE":  "📝 *Ungültiges Befehlsformat*\n\n_Bitte überprüfen Sie die Syntax des Befehls. Antworten Sie mit `!help`, um die Dokumentation anzuzeigen._",
		"ERR_CONTEXT_ERROR":    "⏳ *Sitzung Abgelaufen*\n\n_Der Kontext Ihrer Anfrage konnte nicht ermittelt werden. Bitte geben Sie die Tracking-ID an (z. B. `!edit ABC-1234 ...`)._",
		"MSG_LANG_UPDATED":     "🌐 *Einstellungen Aktualisiert*\n\n_Ihre Schnittstellensprache wurde erfolgreich auf *%s* eingestellt._",
		"MSG_BROADCAST_START":  "📣 *Übertragung Gestartet*\n\n_Ihre Nachricht wird sicher an *%d* autorisierte Gruppen übertragen._",
		"MSG_DELETE_SUCCESS":   "🗑️ *Eintrag Gelöscht*\n\n_Der Logistikeintrag für *%s* wurde sicher entfernt._",
		"MSG_EDIT_SUCCESS":     "✅ *Sendungsdetails Aktualisiert*\n\n🆔 *%s*\n\n📝 *Geänderte Felder:*\n• %s\n\n━━━━━━━━━━━━━━━━━━━━━━━\n_Bitte warten Sie, während wir Ihre aktualisierte digitale Quittung generieren..._",
		"MSG_STATS_HEADER":     "📊 *%s Systemmetriken*",
		"MSG_STATUS_DASHBOARD": "🖥️ *Operations-Dashboard*",
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
