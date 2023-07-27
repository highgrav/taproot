package authn

import "github.com/highgrav/taproot/common"

type UserManager struct {
	UserStore IUserStore
	cache     common.KVCache[User]
}

func (um *UserManager) GetUserById(id string) (User, error) {
	return User{}, nil
}

func (um *UserManager) GetUserByAuth(auth UserAuth) (User, error) {
	return User{}, nil
}
