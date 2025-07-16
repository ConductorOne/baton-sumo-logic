package config

import (
	"github.com/conductorone/baton-sdk/pkg/field"
)

var (
	apiBaseURLField = field.StringField(
		"api-base-url",
		field.WithDisplayName("API base URL"),
		field.WithDescription("The Sumo Logic API base URL. Options include:\n"+
			"- AU: https://api.au.sumologic.com\n"+
			"- CA: https://api.ca.sumologic.com\n"+
			"- DE: https://api.de.sumologic.com\n"+
			"- EU: https://api.eu.sumologic.com\n"+
			"- FED: https://api.fed.sumologic.com\n"+
			"- IN: https://api.in.sumologic.com\n"+
			"- JP: https://api.jp.sumologic.com\n"+
			"- KR: https://api.kr.sumologic.com\n"+
			"- US1 (default): https://api.sumologic.com\n"+
			"- US2: https://api.us2.sumologic.com"),
		field.WithDefaultValue("https://api.sumologic.com"),
	)
	apiAccessIDField = field.StringField(
		"api-access-id",
		field.WithDisplayName("API access ID"),
		field.WithDescription("The Sumo Logic API access ID."),
		field.WithRequired(true),
	)
	apiAccessKeyField = field.StringField(
		"api-access-key",
		field.WithDisplayName("API access key"),
		field.WithDescription("The Sumo Logic API access key."),
		field.WithRequired(true),
		field.WithIsSecret(true),
	)
	includeServiceAccountsField = field.BoolField(
		"include-service-accounts",
		field.WithDisplayName("Include service accounts"),
		field.WithDescription("Whether to include service accounts in the connector."),
		field.WithDefaultValue(true),
	)

	// ConfigurationFields defines the external configuration required for the
	// connector to run. Note: these fields can be marked as optional or
	// required.
	ConfigurationFields = []field.SchemaField{
		apiBaseURLField,
		apiAccessIDField,
		apiAccessKeyField,
		includeServiceAccountsField,
	}

	// FieldRelationships defines relationships between the fields listed in
	// ConfigurationFields that can be automatically validated. For example, a
	// username and password can be required together, or an access token can be
	// marked as mutually exclusive from the username password pair.
	FieldRelationships = []field.SchemaFieldRelationship{}
)

//go:generate go run ./gen
var Config = field.NewConfiguration(
	ConfigurationFields,
	field.WithConstraints(FieldRelationships...),
	field.WithConnectorDisplayName("Sumo Logic"),
	field.WithIconUrl("/static/app-icons/sumo-logic.svg"),
	field.WithHelpUrl("/docs/baton/sumo-logic"),
)
