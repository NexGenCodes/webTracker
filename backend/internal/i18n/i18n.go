package i18n

import (
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
	},
}

func T(lang Language, key string) string {
	l := Language(strings.ToLower(string(lang)))
	if _, ok := Translations[l]; !ok {
		l = EN
	}
	val, ok := Translations[l][key]
	if !ok {
		// Fallback to EN for specific key
		val = Translations[EN][key]
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
