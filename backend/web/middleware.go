package web

import (
	db "backend/db/sqlc"
	"backend/token"
	fsmutil "backend/util/fsm"
	logutil "backend/util/log"
	roleutil "backend/util/role"
	"context"
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	authorizationHeaderKey  = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
	authorizationScopesKey  = "authorization_scopes"
)

func authMiddleware(tokenMaker token.Maker, checkToken func(ctx context.Context, tokenID uuid.UUID) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader(authorizationHeaderKey)

		if len(authorizationHeader) == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, newErrorResponse(codeForbiddenError, messageForbiddenError))
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			c.AbortWithStatusJSON(http.StatusForbidden, newErrorResponse(codeForbiddenError, messageForbiddenError))
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			c.AbortWithStatusJSON(http.StatusForbidden, newErrorResponse(codeForbiddenError, messageForbiddenError))
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, newErrorResponse(codeForbiddenError, messageForbiddenError))
			return
		}

		if ok := checkToken(c, payload.ID); !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, newErrorResponse(codeForbiddenError, messageForbiddenError))
			return
		}

		c.Set(authorizationPayloadKey, payload)
		c.Next()
	}
}

func userScopesMiddleware(s db.IStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

		user, err := s.GetUser(c, authPayload.Subject)
		if err != nil {
			if err == sql.ErrNoRows {
				c.AbortWithStatusJSON(http.StatusForbidden, newErrorResponse(codeForbiddenError, messageForbiddenError))
				return
			}
			logutil.GetLogger().Errorf("get user error, err=%s, user_id=%s", err, authPayload.Subject)
			c.AbortWithStatusJSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
			return
		}

		scopes := roleutil.GetRoleByID(user.RoleID).UserScopes
		c.Set(authorizationScopesKey, scopes)
		c.Next()
	}
}

type storeUserScopesMiddlewareUri struct {
	StoreID *string `uri:"store_id"`
}

func storeUserScopesMiddleware(s db.IStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req storeUserScopesMiddlewareUri
		if err := c.ShouldBindUri(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, messageWrongRequestPayload))
			return
		}

		if req.StoreID == nil || *req.StoreID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, newErrorResponse(codeInvalidParameterError, "store_id is null or empty"))
			return
		}

		storeID, err := uuid.Parse(*req.StoreID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, newErrorResponse(codeForbiddenError, "not a store user"))
			return
		}

		authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

		arg := db.GetStoreUserParams{
			StoreID: storeID,
			UserID:  authPayload.Subject,
		}

		storeUser, err := s.GetStoreUser(c, arg)
		if err != nil {
			if err == sql.ErrNoRows {
				c.AbortWithStatusJSON(http.StatusForbidden, newErrorResponse(codeForbiddenError, "not a store user"))
				return
			}
			logutil.GetLogger().Errorf("get store user error, err=%s, store_id=%s, user_id=%s", err, storeID, authPayload.Subject)
			c.AbortWithStatusJSON(http.StatusInternalServerError, newErrorResponse(codeInternalError, messageServerInternalError))
			return
		}

		if storeUser.State != fsmutil.StoreUserStateActive {
			c.AbortWithStatusJSON(http.StatusForbidden, newErrorResponse(codeForbiddenError, "the store user is diabled"))
			return
		}

		scopes := roleutil.GetRoleByID(storeUser.RoleID).StoreUserScopes
		c.Set(authorizationScopesKey, scopes)
		c.Next()
	}
}

func checkScopesMiddleware(requiredScopesList ...roleutil.Scopes) gin.HandlerFunc {
	return func(c *gin.Context) {
		authScopes := c.MustGet(authorizationScopesKey).(roleutil.Scopes)

		if isAnyScopesMatched(authScopes, requiredScopesList...) {
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusForbidden, newErrorResponse(codeForbiddenError, messageForbiddenError))
			return
		}
	}
}

func isAnyScopesMatched(givenScopes roleutil.Scopes, requiredScopesList ...roleutil.Scopes) bool {
	if len(requiredScopesList) == 0 {
		return true
	}
	for _, scopes := range requiredScopesList {
		if isScopesMatched(givenScopes, scopes) {
			return true
		}
	}
	return false
}

func isScopesMatched(givenScopes roleutil.Scopes, requiredScopes roleutil.Scopes) bool {
	for _, scope := range requiredScopes {
		if !contains(givenScopes, scope) {
			return false
		}
	}
	return true
}

func contains[T comparable](s []T, e T) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
