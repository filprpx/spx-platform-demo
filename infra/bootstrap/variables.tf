variable "api_display_name" {
  type    = string
  default = "SPX Platform API"
}

variable "cli_display_name" {
  type    = string
  default = "SPX Platform CLI"
}

variable "api_identifier_uri" {
  type    = string
  default = "api://spx-platform-api"
}

variable "redirect_uri" {
  type    = string
  default = "http://localhost:8765/callback"
}

variable "owner_object_id" {
  type        = string
  description = "Optional Entra object ID to own both application registrations."
  default     = null
  nullable    = true
}
