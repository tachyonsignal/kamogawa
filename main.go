package main

import (
	"kamogawa/asset"
	"kamogawa/core"
	"kamogawa/handler"
	"kamogawa/identity"
	"kamogawa/types"
	"log"

	"net/http"

	"os"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

func init() {
	shimogawaUrl := os.Getenv("SHIMOGAWA_URL")
	if len(shimogawaUrl) == 0 {
		log.Panic("$SHIMOGAWA_URL not set")
	}
	db = core.InitDB(shimogawaUrl)

	// TODO: remove. for prototyping purposes.
	db.Create(&types.User{
		Email:    "1337gamer@gmail.com",
		Password: "1234",
	})

	// TODO: Use to check if cache works
	//db.Create(&types.Project{
	//	Email:          "1337gamer@gmail.com",
	//	ProjectNumber:  "472481681891",
	//	ProjectId:      "diceduckmonk1",
	//	LifeCycleState: "ACTIVE",
	//	Name:           "diceduckmonk1",
	//	CreateTime:     "2022-08-30T12:30:58.750Z",
	//})
}

func main() {
	r := gin.New()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(identity.SetAuthContext())
	asset.Config(r)

	r.GET("/login", handler.Login)
	r.GET("/reset", handler.Reset)
	r.POST("/login", identity.HandleLogin)
	r.POST("/reset", identity.HandleReset)
	r.GET("/loggedout", func(c *gin.Context) {
		if identity.ExtractClaimsEmail(c) == nil {
			core.HTMLWithGlobalState(c, "loggedout.html", gin.H{})
		} else {
			c.Redirect(http.StatusFound, "/account")
		}
	})

	r.GET("/status", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/login")
	})

	authed := r.Group("/", identity.GateAuth())
	{
		authed.GET("release.txt", asset.TXT(asset.Release))

		// ******** Begin left nav
		authed.GET("/search", handler.Search(db))

		authed.GET("/overview", handler.Overview(db))
		authed.GET("/gce", handler.GCE(db))
		authed.GET("/gae", handler.GAE(db))
		authed.GET("/sql", handler.SQL(db))
		authed.GET("/functions", handler.Functions(db))

		authed.GET("/export", handler.Export)
		authed.GET("/authorization", handler.Authorization(db))
		authed.GET("/account", handler.Account)
		authed.POST("/logout", identity.HandleLogout)
		// ******** End left nav

		authed.GET("/revokegcp", identity.RevokeGCP(db))
		authed.GET("/google/oauth2", identity.GoogleOAuth2Callback(db))

		// Fake privileged routes for demo
		authed.GET("/password", func(c *gin.Context) {
			core.HTMLWithGlobalState(c, "password.html", gin.H{})
			c.Abort()
		})
		authed.GET("/encryption", func(c *gin.Context) {
			core.HTMLWithGlobalState(c, "encryption.html", gin.H{})
			c.Abort()
		})
		authed.GET("/2fa", func(c *gin.Context) {
			core.HTMLWithGlobalState(c, "2fa.html", gin.H{})
			c.Abort()
		})
	}

	r.Run(":3000")
}
