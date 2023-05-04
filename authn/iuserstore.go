package authn

type IUserStore interface {
	GetUserById(id string) (User, error)
	GetUserByAuth(auth UserAuth) (User, error)
	CheckUserRight(userId, domainId, userRight, itemType, itemId string) (bool, error)
}
