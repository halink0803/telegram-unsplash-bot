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
