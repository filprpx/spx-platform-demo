output "tenant_id" {
  value = data.azuread_client_config.current.tenant_id
}

output "api_client_id" {
  value = azuread_application.api.client_id
}

output "api_identifier_uri" {
  value = var.api_identifier_uri
}

output "api_scope" {
  value = "${var.api_identifier_uri}/access_as_user"
}

output "cli_client_id" {
  value = azuread_application.cli.client_id
}

output "redirect_uri" {
  value = var.redirect_uri
}
