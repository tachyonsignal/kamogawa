package handler

import (
	"html/template"
	"kamogawa/core"
	"kamogawa/gcp"
	"kamogawa/identity"
	"kamogawa/types"
	"strings"

	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
)

func Functions(db *gorm.DB) func(*gin.Context) {
	return func(c *gin.Context) {
		useCache := len(c.Query("r")) == 0

		var email, exists = c.Get(identity.IdentityContextkey)
		if !exists {
			panic("Unexpected")
		}
		var user types.User
		err := db.First(&user, "email = ?", email).Error
		if err != nil {
			panic("Unvalid DB state")
		}
		if user.AccessToken == nil {
			core.HTMLWithGlobalState(c, "functions.html", gin.H{
				"Unauthorized": true,
			})
			return
		}
		identity.CheckDBAndRefreshToken(c, user, db)

		responseSuccess, responseError := gcp.GCPListProjects(db, user, useCache)
		if responseError.Error.Code == 403 && strings.HasPrefix(responseError.Error.Message, "Request had insufficient authentication scopes.") {
			core.HTMLWithGlobalState(c, "functions.html", gin.H{
				"MissingScopes": true,
			})
			return
		}

		var projectStrings []types.Project
		if user.Scope == nil || !user.Scope.Valid {
			projectStrings = []types.Project{}
		} else {
			projectStrings = responseSuccess.Projects
		}

		var htmlLines []string
		// Enumerate Projects for credentials
		for _, p := range projectStrings {
			responseSuccess, responseError := gcp.GCFListFunctions(user, p.ProjectId, nil)
			if responseError != nil && responseError.Error.Code > 0 {
				// Shortcircuit on first API call with missing scope to GCF.
				if responseError.Error.Code == 403 && strings.HasPrefix(responseError.Error.Message, "Request had insufficient authentication scopes.") {
					core.HTMLWithGlobalState(c, "functions.html", gin.H{
						"MissingScopes": true,
					})
					return
				}

				if responseError.Error.Code == 403 && strings.HasPrefix(responseError.Error.Message, "Cloud Functions API has not been used in project") {
					htmlLines = append(htmlLines, "<li>Cloud Functions API has not been abled in project</li>")
				} else {
					htmlLines = append(htmlLines, "<li>Error: \""+responseError.Error.Message+"\"</li>")
				}
			} else {
				if len(responseSuccess.Functions) == 0 {
					htmlLines = append(htmlLines, "<li>project"+p.ProjectId+" has no functions</li>")
				} else {
					for _, f := range responseSuccess.Functions {
						htmlLines = append(htmlLines, "<li>id: "+f.Name+"</li>")
					}
				}
			}
		}

		core.HTMLWithGlobalState(c, "functions.html", gin.H{
			"AssetLines": template.HTML(strings.Join(htmlLines[:], "")),
		})
	}
}
