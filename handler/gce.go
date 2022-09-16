package handler

import (
	"database/sql"
	"fmt"
	"html/template"
	"kamogawa/core"
	"kamogawa/gcp"
	"kamogawa/identity"
	"kamogawa/types"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"gorm.io/gorm"
)

func GCE(db *gorm.DB) func(*gin.Context) {
	return func(c *gin.Context) {
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
			core.HTMLWithGlobalState(c, "gce.html", gin.H{
				"Unauthorized": true,
			})
			return
		}

		responseSuccess, responseError := gcp.GCPListProjects(db, user)
		if responseError != nil && responseError.Error.Code == 403 && strings.HasPrefix(responseError.Error.Message, "Request had insufficient authentication scopes.") {
			core.HTMLWithGlobalState(c, "gce.html", gin.H{
				"MissingScopes": true,
			})
			return
		}
		// TODO: this bakes in assumption that responseError not nil IFF 401.
		if responseError != nil && responseError.Error.Code > 0 {
			fmt.Printf("Retrying project fetch by first refreshing access token")
			respRefreshtoken := identity.GCPRefresh(c, db)
			user.AccessToken = &sql.NullString{String: respRefreshtoken.AccessToken, Valid: true}
			db.Save(&user)
			responseSuccess, _ = gcp.GCPListProjects(db, user)
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
			responseSuccess, responseError := gcp.GCEListInstances(db, user, p.ProjectId)
			htmlLines = append(htmlLines, "<li>"+p.ProjectId+" ( Project ) <ul>")
			if responseError.Error.Code > 0 {
				// Shortcircuit if missing GCE scope.
				// TODO: refactor to do oonce utside loop
				if responseError.Error.Code == 403 && strings.HasPrefix(responseError.Error.Message, "Request had insufficient authentication scopes.") {
					core.HTMLWithGlobalState(c, "gce.html", gin.H{
						"MissingScopes": true,
					})
					return
				}

				if responseError.Error.Code == 403 && strings.HasPrefix(responseError.Error.Message, "Compute Engine API has not been used in project") {
					htmlLines = append(htmlLines, "<li>Compute Engine API has not been enabled on project.</li>")
				} else {
					htmlLines = append(htmlLines, "<li>Unknown error with code: "+strconv.Itoa(responseError.Error.Code)+"</li>")
				}
			} else {
				if len(responseSuccess.Zones) == 0 {
					htmlLines = append(htmlLines, "<li>There are no instances in this project.</li>")
				} else {
					for _, zone := range responseSuccess.Zones {
						htmlLines = append(htmlLines, "<li>"+zone.Zone+" ( Zone ) <ul>")
						for _, instance := range zone.Instances {
							htmlLines = append(htmlLines, "<li>"+instance.Name+" ( Instance )</li>")
						}
						htmlLines = append(htmlLines, "</ul></li>")
					}
				}
			}
			htmlLines = append(htmlLines, "</ul>")
		}

		core.HTMLWithGlobalState(c, "gce.html", gin.H{
			"AssetLines": template.HTML(strings.Join(htmlLines[:], "")),
		})
	}
}
