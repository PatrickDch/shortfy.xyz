package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	valid "github.com/asaskevich/govalidator"
	"github.com/boltdb/bolt"
	"github.com/julienschmidt/httprouter"
	"github.com/robfig/cron"
)

var baseUrl = "https://shortfy.xyz/p/" // Replace this url with your server goShort server url
var boltDBPath = "db/shortURL.db"
var shortUrlBkt = []byte("shortUrlBkt")
var ttlBkt = []byte("ttl")
var seedChars = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var seedCharsLen = len(seedChars)
var aChar byte = 97
var dbConn *bolt.DB

type Response struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
	Url    string `json:"url"`
}

type Post struct {
	Url string `json:"url"`
}

func main() {
	var err error
	dbConn, err = bolt.Open(boltDBPath, 0644, nil)
	if err != nil {
		log.Println(err)
	}

	c := cron.New()
	c.AddFunc("0 20 1 * *", func() { fmt.Println(time.Now().UTC().Format(time.RFC3339Nano), "Trying to Sweep") })
	c.AddFunc("0 21 1 * *", func() { Sweep(1) })
	c.Start()

	//defer dbConn.Close()
	router := httprouter.New()
	router.GET("/:code", Redirect)
	router.GET("/:code/json", GetOriginalURL)
	router.POST("/create/", Create)
	fmt.Println("Server started on :8090 port")
	log.Fatal(http.ListenAndServe(":8090", router))

}

func Create(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	//urlStr := r.FormValue("url")
	var po Post
	err := json.NewDecoder(r.Body).Decode(&po)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	urlStr := string(po.Url)
	fmt.Println(string(urlStr))

	//urlStr = strings.Trim(urlStr, " ")
	if valid.IsURL(urlStr) == false && !strings.HasPrefix(urlStr, "magnet:") && !strings.HasPrefix(urlStr, "xml:") {
		resp := &Response{Status: http.StatusBadRequest, Msg: "Invalid input URL", Url: urlStr}
		respJson, _ := json.Marshal(resp)
		fmt.Fprint(w, string(respJson))
		return
	}

	newCode, err := GetNextCode()
	if err != nil {
		resp := Response{Status: http.StatusInternalServerError, Msg: "Some error occured while creating short URL", Url: ""}
		respJson, _ := json.Marshal(resp)
		fmt.Fprint(w, string(respJson))
	}

	byteKey, byteUrl := []byte(newCode), []byte(urlStr)
	err = dbConn.Update(func(tx *bolt.Tx) error {
		//@todo : move this code to main function
		bucket, err := tx.CreateBucketIfNotExists(shortUrlBkt)
		if err != nil {
			return err
		}

		bucketttl, err := tx.CreateBucketIfNotExists(ttlBkt)
		if err != nil {
			return err
		}

		err = bucket.Put(byteKey, byteUrl)
		if err != nil {
			return err
		}

		err = bucketttl.Put([]byte(time.Now().UTC().Format(time.RFC3339Nano)), byteKey)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		log.Println(err)
		resp := &Response{Status: http.StatusInternalServerError, Msg: "Some error occured while creating short URL:", Url: ""}
		respJson, _ := json.Marshal(resp)
		fmt.Fprint(w, string(respJson))
		return
	}

	shortUrl := baseUrl + newCode
	resp := &Response{Status: http.StatusOK, Msg: "Short URL created successfully", Url: shortUrl}
	respJson, _ := json.Marshal(resp)
	fmt.Fprint(w, string(respJson))
}

func Redirect(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	code := ps.ByName("code")
	originalUrl, err := getCodeURL(code)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	http.Redirect(w, r, originalUrl, http.StatusFound)
}

func GetOriginalURL(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	code := ps.ByName("code")
	originalUrl, err := getCodeURL(code)

	if err != nil {
		resp := &Response{Status: http.StatusInternalServerError, Msg: "Some error occured while reading URL", Url: ""}
		respJson, _ := json.Marshal(resp)
		fmt.Fprint(w, string(respJson))
		return
	}

	var resp *Response
	if len(originalUrl) != 0 {
		resp = &Response{Status: http.StatusOK, Msg: "Found", Url: originalUrl}
	} else {
		resp = &Response{Status: http.StatusNotFound, Msg: "URL not found", Url: ""}
	}

	respJson, err := json.Marshal(resp)

	if err != nil {
		fmt.Fprint(w, "Error occurred while creating json response")
		return
	}

	fmt.Fprint(w, string(respJson))
}

func getCodeURL(code string) (string, error) {
	key := []byte(code)
	var originalUrl string

	err := dbConn.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(shortUrlBkt)
		if bucket == nil {
			return fmt.Errorf("Bucket %q not found!", shortUrlBkt)
		}

		value := bucket.Get(key)
		originalUrl = string(value)
		return nil
	})

	if err != nil {
		return "", err
	}
	return originalUrl, nil
}

func GetNextCode() (string, error) {
	var newCode string
	err := dbConn.Update(func(tx *bolt.Tx) error {
		// by using locking on db file BoldDB makes sure it will be thread safe operation
		// and no two goroutines can can get same a short code at a time
		bucket, err := tx.CreateBucketIfNotExists(shortUrlBkt)
		if err != nil {
			return err
		}

		existingCodeByteKey := []byte("existingCodeKey")
		existingCode := bucket.Get(existingCodeByteKey)
		newCode, err = GenerateNextCode(string(existingCode))
		if err != nil {
			return err
		}

		err = bucket.Put(existingCodeByteKey, []byte(newCode))
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return "", err
	}
	return newCode, nil
}

/*
Following method is used to generate alphanumeric incremental code, which will be helpful
for generating short urls
this function will return new code like, input > output
a > b, ax > ay, az > aA, aZ > a1, a9 > ba, 99 > aaa
it will create shortest alphanumeric code possible for using in url
*/
func GenerateNextCode(code string) (string, error) {
	if code == "" {
		return string(aChar), nil
	}
	codeBytes := []byte(code)
	codeByteLen := len(codeBytes)

	codeCharIndex := -1
	for i := (codeByteLen - 1); i >= 0; i-- {
		codeCharIndex = bytes.IndexByte(seedChars, codeBytes[i])
		if codeCharIndex == -1 || codeCharIndex >= seedCharsLen {
			return "", errors.New("Invalid exisitng code")
		} else if codeCharIndex == (seedCharsLen - 1) {
			codeBytes[i] = aChar
		} else {
			codeBytes[i] = seedChars[(codeCharIndex + 1)]
			return string(codeBytes), nil
		}
	}
	for _, byteVal := range codeBytes {
		if byteVal != aChar {
			return string(codeBytes), nil
		}
	}
	// prepending "a" for generating new incremental code
	return "a" + string(codeBytes), nil
}

func Sweep(maxAge time.Duration) (err error) {
	fmt.Println(time.Now(), "Sweep Sweep")
	keys, err := GetExpired(maxAge)
	if err != nil || len(keys) == 0 {
		return
	}

	return dbConn.Update(func(tx *bolt.Tx) (err error) {
		bSecret := tx.Bucket(shortUrlBkt)

		for _, key := range keys {
			if err = bSecret.Delete(key); err != nil {
				return
			}
		}
		fmt.Println("Keys for D", keys)
		return
	})
}

func GetExpired(maxAge time.Duration) (keys [][]byte, err error) {
	keys = [][]byte{}
	ttlKeys := [][]byte{}

	err = dbConn.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(ttlBkt).Cursor()

		max := []byte(time.Now().UTC().Add(-maxAge).Format(time.RFC3339Nano))
		for k, v := c.First(); k != nil && bytes.Compare(k, max) <= 0; k, v = c.Next() {
			keys = append(keys, v)
			ttlKeys = append(ttlKeys, k)
		}
		return nil
	})

	err = dbConn.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(ttlBkt)
		for _, key := range ttlKeys {
			if err = b.Delete(key); err != nil {
				return err
			}
		}
		return nil
	})

	return
}
