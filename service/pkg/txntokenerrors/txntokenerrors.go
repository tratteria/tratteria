package txntokenerrors

import (
	"errors"
)

var ErrParsingSubjectToken = errors.New("error parsing subject token")

var ErrInvalidSubjectTokenClaims = errors.New("invalid subject token claims")

var ErrUnsupportedTokenType = errors.New("token type not supported")

var ErrConfiguredSubjectFieldNotFound = errors.New("configured subject field not found in the subject token")
