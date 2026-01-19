package olympic

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

var database *DB
var accountManager *AccountManager
var linkManager *LinkManager

func createAccount(c *gin.Context) {
	c.Request.ParseForm()
	email := c.Request.Form.Get("email")
	password := c.Request.Form.Get("password")
	_, err := accountManager.CreateAccount(email, password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
		})
		log.Println("CreateAccount error:", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func login(c *gin.Context) {
	c.Request.ParseForm()
	email := c.Request.Form.Get("email")
	password := c.Request.Form.Get("password")
	refreshToken, jwtToken, err := accountManager.Login(email, password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
		})
		log.Println("Authentication error:", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":       "ok",
		"refreshToken": refreshToken, // Return refresh token
		"jwtToken":     jwtToken,     // Return JWT token
	})
}

func createLink(c *gin.Context) {
	c.Request.ParseForm()
	accountID, err := checkJWTMiddleware(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
		})
		log.Println("JWT validation error:", err)
		return
	}
	url := c.Request.Form.Get("url")
	shortUrl := c.Request.Form.Get("shortUrl")
	err = linkManager.CreateLink(url, shortUrl, accountID)
	statusCode := http.StatusBadRequest
	if err == ErrShortUrlExists {
		statusCode = http.StatusConflict
	}
	if err != nil {
		c.JSON(statusCode, gin.H{
			"status": "error",
		})
		log.Println("CreateLink error:", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"url":      url,
		"shortUrl": shortUrl,
	})
}

func getLink(c *gin.Context) {
	c.Request.ParseForm()
	_, err := checkJWTMiddleware(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
		})
		log.Println("JWT validation error:", err)
		return
	}
	shortUrl := c.Request.Form.Get("shortUrl")
	url, err := linkManager.GetLink(shortUrl)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
		})
		log.Println("GetLink error:", err)
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func refreshJWT(c *gin.Context) {
	c.Request.ParseForm()
	refreshToken := c.Request.Form.Get("refreshToken")
	login, err := accountManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
		})
		log.Println("RefreshJWT validation error:", err)
		return
	}
	newJWT, err := accountManager.GenerateJWT(login.AccountID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
		})
		log.Println("RefreshJWT error:", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"jwtToken": newJWT,
	})
}

func logout(c *gin.Context) {
	c.Request.ParseForm()
	refreshToken := c.Request.Form.Get("refreshToken")
	login, err := accountManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
		})
		log.Println("Logout error:", err)
		return
	}
	err = accountManager.Logout(login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
		})
		log.Println("Logout error:", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func checkJWTMiddleware(c *gin.Context) (uint, error) {
	jwtToken := c.Request.Form.Get("jwtToken")
	accountID, err := accountManager.ValidateJWT(jwtToken)
	if err != nil {
		log.Println("JWT validation error:", err)
		return 0, err
	}
	return accountID, nil
}

func registerRoutes(r *gin.Engine) {
	// Define a simple GET endpoint
	r.GET("/api/createAccount", createAccount)
	r.GET("/api/createLink", createLink)
	r.GET("/api/getLink", getLink)
	r.GET("/api/login", login)
	r.GET("/api/refreshJWT", refreshJWT)
	r.GET("/api/logout", logout)
}

func Init(dbName string, jwtSigningSecret string) {
	var err error
	database, err = OpenDB(dbName)
	if err != nil {
		log.Panicf("failed to connect database: %v", err)
	}
	log.Println("Connected to Database:", dbName)
	accountManager = &AccountManager{db: database.DB, jwtSigningSecret: jwtSigningSecret}
	linkManager = &LinkManager{db: database.DB}
}

func Run(port int) {
	// Set Gin to release mode to disable debug output
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	registerRoutes(r)
	log.Printf("Starting server on port %d...", port)
	r.Run(":" + fmt.Sprintf("%d", port))
}

func Shutdown() {
	if database != nil {
		err := database.Close()
		if err != nil {
			log.Printf("Error closing database: %v", err)
		} else {
			log.Println("Database connection closed.")
		}
	}
}
