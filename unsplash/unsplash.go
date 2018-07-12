package unsplash

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/halink0803/telegram-unsplash-bot/common"
)

const endpoint string = "https://api.unsplash.com/"

//int - user id, string - access token
var bearerToken map[int]string

//AuthorizeResp reponse from authorize api
type AuthorizeResp struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	CreatedAt   int    `json:"created_at"`
}

//Unsplash object which interact with Unsplash using Unsplash api
type Unsplash struct {
	client         http.Client
	unsplashKey    string
	unsplashSecret string
}

//NewUnsplash create a new unsplash instance
func NewUnsplash(unsplashKey, unsplashSecret string) *Unsplash {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	return &Unsplash{
		client:         client,
		unsplashKey:    unsplashKey,
		unsplashSecret: unsplashSecret,
	}
}

//AuthorizeUser authorize user
func (u Unsplash) AuthorizeUser(code string, userID int) error {
	reqURL := fmt.Sprintf("%s/oauth/token", "https://unsplash.com")
	req, err := http.NewRequest("POST", reqURL, nil)
	if err != nil {
		log.Panic(err)
	}
	q := req.URL.Query()
	q.Add("client_id", u.unsplashKey)
	q.Add("client_secret", u.unsplashSecret)
	q.Add("redirect_uri", "urn:ietf:wg:oauth:2.0:oob")
	q.Add("code", code)
	q.Add("grant_type", "authorization_code")
	req.Header.Add("Accept", "application/json")
	req.URL.RawQuery = q.Encode()

	log.Printf("request: %+v", req)
	client := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Panic(err)
	}
	var respBody []byte
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			log.Printf("Response body close error: %s", cErr.Error())
		}
	}()
	respBody, err = ioutil.ReadAll(resp.Body)
	var response AuthorizeResp
	log.Printf("%s", respBody)
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return err
	}
	if len(bearerToken) == 0 {
		bearerToken = map[int]string{}
	}
	bearerToken[userID] = response.AccessToken
	log.Printf("%+v", bearerToken)
	return err
}

//UnsplashKey Expose unsplash key
func (u Unsplash) UnsplashKey() string {
	return u.unsplashKey
}

//LikeAPhoto like a photo in unsplash
func (u Unsplash) LikeAPhoto(photoID string) error {
	requestURL := fmt.Sprintf("%s/photos/%s/like", endpoint, photoID)
	contentType := "application/x-www-form-urlencoded"
	resp, err := u.client.Post(requestURL, contentType, nil)
	if err != nil {
		return err
	}
	log.Printf("Like response: %+v", resp)
	return nil
}

//UnlikeAPhoto unlike a photo if its previously liked
func (u Unsplash) UnlikeAPhoto() error {
	//TODO: implement it
	return nil
}

//SearchPhotos return photo url
func (u Unsplash) SearchPhotos(query string) (common.SearchResult, error) {
	requestURL := fmt.Sprintf("%s/search/photos?query=%s&client_id=%s", endpoint, query, u.unsplashKey)
	resp, err := u.client.Get(requestURL)
	result := common.SearchResult{}
	if err != nil {
		return result, err
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	json.Unmarshal(respBody, &result)
	return result, nil
}

//DownloadAPhoto download a photo from unsplash
func (u Unsplash) DownloadAPhoto(photoID string) {
	// TODO: get download link for a photo from unsplash
}
