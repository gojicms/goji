package sessions

import (
	"errors"
	"net/http"
	"time"

	"github.com/gojicms/goji/core/config"
	"github.com/gojicms/goji/core/database"
	"github.com/gojicms/goji/core/extend"
	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/services/auth/users"
	"github.com/gojicms/goji/core/utils"
	"github.com/gojicms/goji/core/utils/log"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

//////////////////////////////////
// Types Definitions            //
//////////////////////////////////

type Session struct {
	gorm.Model
	SessionId string    `gorm:"index;size:36"` // UUID is 36 chars
	CSRF      string    `gorm:"size:255"`
	UserId    uint      `gorm:"index"`
	ExpiresAt time.Time `gorm:"index"`
}

var Service = extend.ServiceDef{
	Name:         "sessions",
	FriendlyName: "Sessions",
	Internal:     true,
	Resources:    []extend.ResourceDef{},
	OnInit: func() error {
		CleanUpSessions()

		db := database.GetDB()
		_ = db.AutoMigrate(&Session{})

		extend.AddMiddleware(extend.NewMiddleware("*", "*", 0, func(flow *httpflow.HttpFlow) {
			var user *users.User
			session, _ := EnsureSession(flow)

			if session != nil {
				user, _ = users.GetById(session.UserId)
				flow.Set("session", session)
				flow.Set("user", user)
				flow.Append("templateData", "user", user)
			}
		}))
		return nil
	},
}

//////////////////////////////////
// Public  Methods              //
//////////////////////////////////

func IsAuthenticated(r *http.Request) bool {
	session, _ := GetSessionFromRequest(r)
	return session != nil
}

func EnsureSession(flow *httpflow.HttpFlow) (*Session, error) {
	session, _ := GetSessionFromRequest(flow.Request)

	if session == nil {
		return nil, errors.New("session not found")
	}

	refreshLifetime := config.ActiveConfig.Application.Auth.RefreshLifetime
	cookieLifetime := config.ActiveConfig.Application.Auth.CookieLifetime
	refreshTime := session.ExpiresAt.Add(-cookieLifetime).Add(refreshLifetime)

	log.Debug("Sessions", "Session set to expire at %s", session.ExpiresAt.Format(time.RFC3339))
	log.Debug("Sessions", "Session set to renew at %s", refreshTime.Format(time.RFC3339))

	// If the session isn't exired but is past the refresh point, renew.
	if time.Now().After(refreshTime) {
		log.Debug("Sessions", "Session Expired - Renewing Session")

		sessionId := uuid.New().String()
		expiration := time.Now().Add(config.ActiveConfig.Application.Auth.CookieLifetime)
		session.SessionId = sessionId
		session.ExpiresAt = expiration

		db := database.GetDB()
		db.Model(&session).Save(session)

		flow.SetCookie(&http.Cookie{
			Name:     config.ActiveConfig.Application.Auth.CookieId,
			Value:    sessionId,
			Expires:  expiration,
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   true,
			Path:     "/",
		})

		// Redirect to the same page to reload
		flow.Redirect(flow.Request.URL.String(), http.StatusFound)
	}

	// Uh oh - this session should not exist
	if getUserForSession(session.SessionId) == nil {
		deleteSessionById(session.SessionId)
		return nil, nil
	}

	return session, nil
}

func CreateSession(flow *httpflow.HttpFlow, csrf string, userId uint) (*Session, error) {
	expiration := time.Now().Add(config.ActiveConfig.Application.Auth.CookieLifetime)
	sessionID := uuid.New().String()

	session := Session{
		SessionId: sessionID,
		ExpiresAt: expiration,
		UserId:    userId,
		CSRF:      csrf,
	}

	db := database.GetDB()
	err := db.Create(&session).Error
	if err != nil {
		return nil, err
	}

	flow.SetCookie(&http.Cookie{
		Name:     "_CSRF",
		Value:    csrf,
		Expires:  time.Now().Add(time.Hour * 24 * 14),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
		Path:     "/",
	})

	flow.SetCookie(&http.Cookie{
		Name:     config.ActiveConfig.Application.Auth.CookieId,
		Value:    sessionID,
		Expires:  expiration,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   true,
		Path:     "/",
	})

	return &session, nil
}

func EndSession(flow *httpflow.HttpFlow) {
	session := flow.Get("session")
	if session == nil {
		return
	}

	db := database.GetDB()
	db.Unscoped().Delete(&Session{}, session)

	// Expire the cookie
	flow.SetCookie(&http.Cookie{
		Name:     config.ActiveConfig.Application.Auth.CookieId,
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
}

func GetAllSessions() []Session {
	db := database.GetDB()
	var sessions []Session
	db.Find(&sessions)
	return sessions
}

func GetSessionFromRequest(r *http.Request) (*Session, error) {
	authCookie, err := r.Cookie(config.ActiveConfig.Application.Auth.CookieId)
	if err != nil {
		return nil, err
	}

	session := findSessionById(authCookie.Value)
	return session, nil
}

func CleanUpSessions() {
	db := database.GetDB()
	db.Unscoped().Where("expires_at < ?", time.Now()).Delete(&Session{})
}

//////////////////////////////////
// Private Methods              //
//////////////////////////////////

func findSessionById(id string) *Session {
	db := database.GetDB()
	var session = Session{}
	db.Model(&session).First(&session, utils.Object{"session_id": id})
	if session.SessionId == "" || session.ExpiresAt.Before(time.Now()) {
		return nil
	}
	return &session
}

func getUserForSession(sessionId string) *users.User {
	db := database.GetDB()
	var user users.User

	session := findSessionById(sessionId)

	db.Model(&user).First(&user, session.UserId)
	return &user
}

func deleteSessionById(sessionId string) {
	db := database.GetDB()
	db.Delete(&Session{}, sessionId)
}
