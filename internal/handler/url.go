package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"go-shorten/internal/model"
	"go-shorten/internal/repository"
	"go-shorten/internal/store"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	URL_CODE_LENGTH        = 8
	CUSTOM_CODE_MIN_LENGTH = 4
	CUSTOM_CODE_MAX_LENGTH = URL_CODE_LENGTH * 2
)

var (
	ErrNoCodeSpecified    = errors.New("no code supplied")
	ErrCodeAlreadyUsed    = errors.New("code already used")
	ErrCustomCodeTooShort = errors.New("custom code too short")
	ErrCustomCodeTooLong  = errors.New("custom code too long")
	ErrDaysOutOfRange     = errors.New("no of days should be between 1 to 7")
)

func ShortenURLHandler(c *gin.Context) {
	var url model.URL
	if err := c.ShouldBindJSON(&url); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if url.Days < 1 || url.Days > 7 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": ErrDaysOutOfRange.Error()})
		return
	}

	var encodedUrl string
	var tries uint
	for url.CustomCode == "" {
		tries++

		h := sha256.New()
		h.Write([]byte(rand.Text()))
		h.Write([]byte(url.URL))

		encodedUrl = base64.RawURLEncoding.EncodeToString(h.Sum(nil))[:URL_CODE_LENGTH]
		if _, ok := store.GetFromCache(encodedUrl); ok {
			continue
		}

		// Is the code generation algorithm random enough?
		if exists, err := repository.CheckCodeExists(c, encodedUrl); err != nil {
			log.Println("Error: ", err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else if exists == 0 {
			log.Println("Took", tries, "tries to generate code")
			break
		}
	}

	if l := len(url.CustomCode); l > CUSTOM_CODE_MAX_LENGTH {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": ErrCustomCodeTooLong.Error()})
		return
	} else if l < CUSTOM_CODE_MIN_LENGTH {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": ErrCustomCodeTooShort.Error()})
		return
	} else {
		if _, ok := store.GetFromCache(url.CustomCode); ok {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": ErrCodeAlreadyUsed.Error()})
			return
		}

		if exists, err := repository.CheckCodeExists(c, url.CustomCode); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else if exists == 1 {
			store.AddToCache(encodedUrl, url.URL)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": ErrCodeAlreadyUsed.Error()})
			return
		} else {
			encodedUrl = url.CustomCode
		}

		if err := repository.AddShortenedUrl(c, url.URL, encodedUrl, url.Days); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		store.AddToCache(encodedUrl, url.URL)
		c.JSON(http.StatusCreated, gin.H{"code": encodedUrl})
	}
}

func RetrieveMappingHandler(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if url, ok := store.GetFromCache(code); ok {
		c.Redirect(http.StatusFound, url)
		return
	}

	if url, err := repository.GetShortenedURL(c, code); err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
	} else {
		store.AddToCache(code, url)
		c.Redirect(http.StatusFound, url)
	}
}
