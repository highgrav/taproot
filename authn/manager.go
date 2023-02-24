package authn

import "highgrav/taproot/v1/common"

type UserManager struct {
	UserStore IUserStore
	cache     common.KVCache[User]
}
