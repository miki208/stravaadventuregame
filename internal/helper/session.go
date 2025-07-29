package helper

import (
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type Session struct {
	UserId        int64
	SessionCookie http.Cookie
}

type SessionManager struct {
	sessionIdToSession map[string]Session
	userIdToSessionId  map[int64]string
}

func CreateSessionManager() *SessionManager {
	sessionManager := &SessionManager{
		sessionIdToSession: make(map[string]Session),
		userIdToSessionId:  make(map[int64]string),
	}

	return sessionManager
}

func (manager *SessionManager) GetSessionBySessionId(sessionId string) *Session {
	session, ok := manager.sessionIdToSession[sessionId]
	if !ok {
		return nil
	}

	return &session
}

func (manager *SessionManager) GetSessionByUserId(userId int64) *Session {
	sessionId, ok := manager.userIdToSessionId[userId]
	if !ok {
		return nil
	}

	return manager.GetSessionBySessionId(sessionId)
}

func (manager *SessionManager) GetSessionByRequest(req *http.Request) *Session {
	sessionCookie, err := req.Cookie("session_id")
	if errors.Is(err, http.ErrNoCookie) {
		return nil
	}

	session := manager.GetSessionBySessionId(sessionCookie.Value)
	if session == nil {
		return nil
	}

	if time.Now().After(session.SessionCookie.Expires) {
		manager.DestroySession(*session)

		return nil
	}

	return session
}

func (manager *SessionManager) DestroySession(session Session) {
	sessionId, ok := manager.userIdToSessionId[session.UserId]
	if !ok {
		return
	}

	delete(manager.userIdToSessionId, session.UserId)
	delete(manager.sessionIdToSession, sessionId)

	slog.Debug("Session destroyed.", "userId", session.UserId, "sessionId", sessionId)
}

func (manager *SessionManager) CreateSession(userId int64) Session {
	session := manager.GetSessionByUserId(userId)
	if session != nil {
		manager.DestroySession(*session)
	}

	sessionCookie := CreateSessionCookie()

	manager.userIdToSessionId[userId] = sessionCookie.Value
	manager.sessionIdToSession[sessionCookie.Value] = Session{UserId: userId, SessionCookie: sessionCookie}

	slog.Debug("Session created.", "userId", userId, "sessionId", sessionCookie.Value)

	return manager.sessionIdToSession[sessionCookie.Value]
}
