package pkg

import "errors"

var (
	ErrInvalidCredentials      = errors.New("errors: invalid credentials")
	ErrAccountInActive         = errors.New("errors: account is inactive")
	ErrDuplicateEmail          = errors.New("errors: duplicate email")
	ErrNoRecord                = errors.New("errors: no matching record found")
	ErrNoCookieFound           = errors.New("errors: no cookie found")
	ErrCantFindProduct         = errors.New("errors: can't find product")
	ErrCantDecodeProducts      = errors.New("errors: can't find product")
	ErrUserIDIsNotValid        = errors.New("errors: user is not valid")
	ErrCantAddInCart           = errors.New("errors: cannot add product to cart")
	ErrCantAddUser             = errors.New("errors: cannot add user")
	ErrCantAddProduct          = errors.New("errors: cannot add product")
	ErrCantRemoveItem          = errors.New("errors: cannot remove item from cart")
	ErrCantGetItem             = errors.New("errors: cannot get item from cart ")
	ErrCantBuyCartItem         = errors.New("errors: cannot update the purchase")
	ErrNoEnvFileFound          = errors.New("errors: no matching .env file found")
	ErrIncorrectPassword       = errors.New("errors: incorrect password")
	ErrCantUseGeneratePassword = errors.New("errors: bycrypt function not generating password")
	ErrUserNotFound            = errors.New("errors: no such user exist")
	ErrInvalidUserFound        = errors.New("errors: user access denied")
	ErrInternalServer          = errors.New("errors: internal server error")
)
