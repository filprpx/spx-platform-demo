data "azuread_client_config" "current" {}

locals {
  owners = var.owner_object_id == null ? [data.azuread_client_config.current.object_id] : [var.owner_object_id]
}

resource "azurerm_resource_group" "bootstrap" {
  name     = var.bootstrap_resource_group_name
  location = var.bootstrap_location

  tags = {
    managed_by = "spx-demo"
    purpose    = "bootstrap"
  }
}

resource "random_string" "acr_suffix" {
  length  = 8
  special = false
  upper   = false
}

resource "azurerm_container_registry" "bootstrap" {
  name                = "spxdemo${random_string.acr_suffix.result}"
  resource_group_name = azurerm_resource_group.bootstrap.name
  location            = azurerm_resource_group.bootstrap.location
  sku                 = "Basic"
  admin_enabled       = false

  tags = {
    managed_by = "spx-demo"
    purpose    = "shared-image-registry"
  }
}

resource "azuread_application" "api" {
  display_name     = var.api_display_name
  owners           = local.owners
  sign_in_audience = "AzureADMyOrg"
  identifier_uris  = [var.api_identifier_uri]

  api {
    requested_access_token_version = 2

    oauth2_permission_scope {
      admin_consent_description  = "Allow the CLI to access the SPX platform API on behalf of the signed-in user."
      admin_consent_display_name = "Access SPX Platform API"
      enabled                    = true
      id                         = "b5b1f8c7-8d34-4f2e-b8b3-9a64ad2c6e01"
      type                       = "User"
      user_consent_description   = "Allow the CLI to access the SPX platform API on your behalf."
      user_consent_display_name  = "Access SPX Platform API"
      value                      = "access_as_user"
    }
  }
}

resource "azuread_service_principal" "api" {
  client_id = azuread_application.api.client_id
  owners    = local.owners
}

resource "azuread_application" "cli" {
  display_name     = var.cli_display_name
  owners           = local.owners
  sign_in_audience = "AzureADMyOrg"

  public_client {
    redirect_uris = [var.redirect_uri]
  }

  required_resource_access {
    resource_app_id = azuread_application.api.client_id

    resource_access {
      id   = azuread_application.api.oauth2_permission_scope_ids["access_as_user"]
      type = "Scope"
    }
  }
}

resource "azuread_service_principal" "cli" {
  client_id = azuread_application.cli.client_id
  owners    = local.owners
}
