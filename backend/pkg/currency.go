package pkg

import "simple_bank/constants"

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case constants.CNY, constants.USD, constants.CAD:
		return true
	}
	return false
}
