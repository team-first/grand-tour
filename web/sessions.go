package web

import (
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
	"github.com/team-first/grand-tour/core"
	"net/http"
)

const sessionName = "grand-tour"

var store *sessions.CookieStore

func InitStore(secret []byte) {
	store = sessions.NewCookieStore(secret)
}

func cachedGet(r *http.Request, key interface{}, f func() (interface{}, error)) (value interface{}, err error) {
	value, ok := context.GetOk(r, key)

	if !ok {
		value, err = f()

		if err == nil {
			context.Set(r, key, value)
		}
	}

	return value, err
}

func GetCurrentUser(r *http.Request) (user *core.User, err error) {
	value, err := cachedGet(r, "user", func() (interface{}, error) {
		session, err := store.Get(r, sessionName)

		if err != nil {
			return user, err
		}

		id, ok := session.Values["id"].(int64)

		if ok {
			user = new(core.User)
			user.Id = id
		}

		return user, err
	})

	if err == nil {
		user = value.(*core.User)
	}

	return user, err
}

func Login(w http.ResponseWriter, r *http.Request, user *core.User) (err error) {
	session, err := store.Get(r, sessionName)

	if err != nil {
		return err
	}

	session.Values["id"] = user.Id
	session.Save(r, w)

	context.Set(r, "user", user)

	return err
}

func Logout(w http.ResponseWriter, r *http.Request) (err error) {
	session, err := store.Get(r, sessionName)

	if err != nil {
		return err
	}

	delete(session.Values, "id")
	session.Save(r, w)

	context.Set(r, "user", nil)

	return err
}
