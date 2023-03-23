package authn

import (
	"bufio"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strings"
	"sync"
)

type DigestUserEntry struct {
	Username   string
	Realm      string
	Domain     string
	PwdHash    string
	Workgroups map[string]string
	Labels     []string
}

func NewDigestUserEntry(username, realm, domain, pwd string, workgroups map[string]string, labels []string) (DigestUserEntry, error) {
	due := DigestUserEntry{
		Username:   username,
		Realm:      realm,
		Domain:     domain,
		PwdHash:    "",
		Workgroups: workgroups,
		Labels:     labels,
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return due, err
	}
	due.PwdHash = string(hash)
	return due, nil
}

/*
PasswordFileManager is a trivial user repository using a modified .htdigest file format.
The file format is USERNAME\tREALM\tDOMAIN\tBCRYPT_PWD\t[[WORKGROUP_ID,WORKGROUP_NAME]...]\t[LABEL,...]
Users are naively stored in a simple array; migrating them to the session manager would be a
better idea -- though you shouldn't use this as a large-scale user store in any case.
*/

type PasswordFileManager struct {
	sync.Mutex
	filename string
	users    []DigestUserEntry
}

func NewPasswordFileManager(filename string) (*PasswordFileManager, error) {
	pm := &PasswordFileManager{
		Mutex:    sync.Mutex{},
		filename: filename,
		users:    make([]DigestUserEntry, 0),
	}

	s, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}
	if s.IsDir() {
		return nil, errors.New("htdigest file '" + filename + "' is a directory")
	}
	pm.users, err = pm.readFile()
	if err != nil {
		return nil, err
	}
	return pm, nil
}

func (pfm *PasswordFileManager) readFile() ([]DigestUserEntry, error) {
	pfm.Lock()
	defer pfm.Unlock()
	arr := make([]DigestUserEntry, 0)
	file, err := os.Open(pfm.filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		due, err := readLine(scanner.Text())
		if err != nil {
			return arr, err
		}
		arr = append(arr, due)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return arr, nil
}

func createLine(due DigestUserEntry) string {
	s := strings.Builder{}

	s.Write([]byte(fmt.Sprintf("%s\t%s\t%s\t%s\t", due.Username, due.Realm, due.Domain, due.PwdHash)))

	kvs := ""
	for k, v := range due.Workgroups {
		kvs = kvs + fmt.Sprintf("%s,%s,", k, v)
	}
	s.Write([]byte(kvs[:len(kvs)-1]))

	lbls := ""
	for _, v := range due.Labels {
		lbls = lbls + v + ","
	}
	s.Write([]byte(lbls[:len(lbls)-1]))
	s.Write([]byte("\n"))
	return s.String()
}

func readLine(line string) (DigestUserEntry, error) {
	due := DigestUserEntry{}
	elems := strings.Split(line, "\t")
	if len(elems) < 4 {
		return due, errors.New("malformed line")
	}
	due.Username = elems[0]
	due.Realm = elems[1]
	due.Domain = elems[2]
	due.PwdHash = elems[3]
	due.Labels = make([]string, 0)
	due.Workgroups = make(map[string]string)
	if len(elems) > 4 {
		wgs := strings.Split(elems[4], ",")
		if len(wgs)%2 != 0 {
			return due, errors.New("malformed workgroup section")
		}
		x := 0
		for x = 0; x < len(wgs); {
			due.Workgroups[wgs[x]] = wgs[x+1]
			x = x + 2
		}
	}
	if len(elems) > 5 {
		labels := strings.Split(elems[5], ",")
		for _, l := range labels {
			due.Labels = append(due.Labels, l)
		}
	}
	return due, nil
}

func (pfm *PasswordFileManager) AddNewUser(user DigestUserEntry) error {
	line := createLine(user)
	pfm.Lock()
	defer pfm.Unlock()
	f, err := os.OpenFile(pfm.filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err = f.WriteString(line); err != nil {
		panic(err)
	}
	return nil
}

func (pfm *PasswordFileManager) UpdateUser(user DigestUserEntry) error {
	pfm.Lock()
	defer pfm.Unlock()
	return nil
}

func (pfm *PasswordFileManager) DeleteUser(user DigestUserEntry) error {
	pfm.Lock()
	defer pfm.Unlock()
	return nil
}

func (pfm *PasswordFileManager) GetUserById(realm, id string) (User, error) {
	for _, u := range pfm.users {
		if strings.ToLower(u.Username) == strings.ToLower(id) && strings.ToLower(u.Realm) == strings.ToLower(realm) {
			usr := userFromEntry(u)
			return usr, nil
		}
	}
	return User{}, ErrUserNotFound
}

func (pfm *PasswordFileManager) GetUserByAuth(auth UserAuth) (User, error) {
	if auth.AuthType != AUTH_BASIC {
		return User{}, ErrAuthUnknownScheme
	}
	for _, u := range pfm.users {
		if strings.ToLower(auth.Realm) == strings.ToLower(u.Realm) && strings.ToLower(auth.UserIdentifier) == strings.ToLower(u.Username) {
			if bcrypt.CompareHashAndPassword([]byte(u.PwdHash), []byte(auth.PasswordOrToken)) == nil {
				return userFromEntry(u), nil
			} else {
				return User{}, ErrUserNotAuthenticated
			}
		}
	}
	return User{}, ErrUserNotAuthenticated
}

func userFromEntry(u DigestUserEntry) User {
	usr := User{
		RealmID:                u.Realm,
		UserID:                 u.Username,
		Username:               u.Username,
		DisplayName:            u.Username,
		Emails:                 []string{},
		Phones:                 []string{},
		IsVerified:             true,
		IsBlocked:              false,
		IsActive:               true,
		IsDeleted:              false,
		RequiresPasswordUpdate: false,
		Domains:                []string{u.Domain},
		Workgroups:             WorkgroupMembership{},
		Labels:                 DomainAssertions{},
		Keys:                   nil,
	}
	for k, v := range u.Workgroups {
		usr.Workgroups.AddWorkgroup(u.Domain, k, v)
	}
	usr.Labels[u.Domain] = []string{}
	for _, l := range u.Labels {
		usr.Labels[u.Domain] = append(usr.Labels[u.Domain], l)
	}
	return usr
}
