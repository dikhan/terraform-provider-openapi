package openapi

// Definition level extensions
const extTfImmutable = "x-terraform-immutable"
const extTfForceNew = "x-terraform-force-new"
const extTfSensitive = "x-terraform-sensitive"
const extTfFieldName = "x-terraform-field-name"
const extTfFieldStatus = "x-terraform-field-status"
const extTfID = "x-terraform-id"
const extTfComputed = "x-terraform-computed"
const extTfIgnoreOrder = "x-terraform-ignore-order"
const extIgnoreOrder = "x-ignore-order"

// Operation level extensions
const extTfResourceTimeout = "x-terraform-resource-timeout"
const extTfResourcePollEnabled = "x-terraform-resource-poll-enabled"
const extTfResourcePollTargetStatuses = "x-terraform-resource-poll-completed-statuses"
const extTfResourcePollPendingStatuses = "x-terraform-resource-poll-pending-statuses"
const extTfExcludeResource = "x-terraform-exclude-resource"
const extTfResourceName = "x-terraform-resource-name"
const extTfResourceURL = "x-terraform-resource-host"

// Param level extensions
const extTfHeader = "x-terraform-header"

// Security level extensions
const extTfAuthenticationSchemeBearer = "x-terraform-authentication-scheme-bearer"
const extTfAuthenticationRefreshToken = "x-terraform-refresh-token-url" // #nosec G101
