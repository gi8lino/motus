// Package sounds exposes the shared domain sound utilities via the service layer.
package sounds

import domain "github.com/gi8lino/motus/internal/domain/sounds"

// Option aliases the domain sound option structure.
type Option = domain.Option

// BuiltinOptions exposes the curated sound catalog.
var BuiltinOptions = domain.BuiltinOptions
