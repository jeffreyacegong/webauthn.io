package session

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"net/http"

	log "github.com/duo-labs/webauthn.io/logger"
	"github.com/duo-labs/webauthn/webauthn"
	"github.com/gorilla/sessions"
)

const DefaultEncryptionKeyLength = 32
const WebauthnSession = "webauthn-session"

var ErrInsufficientBytesRead = errors.New("insufficient bytes read")
var ErrMarshal = errors.New("error unmarshaling data")

// GenerateSecureKey reads and returns n bytes from the crypto/rand reader
func GenerateSecureKey(n int) ([]byte, error) {
	buf := make([]byte, n)
	read, err := rand.Read(buf)
	if err != nil {
		return buf, err
	}
	if read != n {
		return buf, ErrInsufficientBytesRead
	}
	return buf, nil
}

type Store struct {
	*sessions.CookieStore
}

func NewStore(keyPairs ...[]byte) (*Store, error) {
	// Generate a default encryption key if one isn't provided
	if len(keyPairs) == 0 {
		key, err := GenerateSecureKey(DefaultEncryptionKeyLength)
		if err != nil {
			return nil, err
		}
		keyPairs = append(keyPairs, key)
	}
	store := &Store{
		sessions.NewCookieStore(keyPairs...),
	}
	return store, nil
}

// SaveWebauthnSession marhsals and saves the webauthn data to the provided
// key given the request and responsewriter
func (store *Store) SaveWebauthnSession(key string, data *webauthn.SessionData, r *http.Request, w http.ResponseWriter) error {
	marshaledData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return store.Set(key, marshaledData, r, w)
}

func (store *Store) GetWebauthnSession(key string, r *http.Request) (webauthn.SessionData, error) {
	sessionData := webauthn.SessionData{}
	session, err := store.Get(r, WebauthnSession)
	if err != nil {
		return sessionData, err
	}
	assertion, ok := session.Values[key].([]byte)
	if !ok {
		return sessionData, ErrMarshal
	}
	err = json.Unmarshal(assertion, &sessionData)
	if err != nil {
		return sessionData, err
	}
	// Delete the value from the session now that it's been read
	delete(session.Values, key)
	return sessionData, nil
}

func (store *Store) Set(key string, value interface{}, r *http.Request, w http.ResponseWriter) error {
	session, err := store.Get(r, WebauthnSession)
	if err != nil {
		log.Errorf("Error getting session %s", err)
	}
	session.Values[key] = value
	session.Save(r, w)
	return nil
}