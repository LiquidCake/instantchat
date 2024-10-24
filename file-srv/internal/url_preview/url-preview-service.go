package url_preview

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/disintegration/imaging"
	"instantchat.rooms/instantchat/file-srv/internal/util"
)

// url preview response format
type UrlPreviewInfo struct {
	Url         *string `json:"u,omitempty"` //url of page
	Title       *string `json:"t,omitempty"`
	Description *string `json:"d,omitempty"`
	ImageUrl    *string `json:"i,omitempty"` //url of og:image from webpage

	ImageDataBase64 *string `json:"b64,omitempty"` //base64 encoded data of image, in case of image URL
}

type UrlPreviewInfoCacheItem struct {
	urlLock            sync.Mutex
	lastCheckTimestamp int64
	lastErrorTimestamp int64
	urlPreviewInfoData UrlPreviewInfo
}

/* Constants */

const UrlToPreviewInfoCacheTTL = 30 * time.Minute
const UrlToPreviewErrorReCheckDelay = 3 * time.Minute

const ClearOldCacheItemsFuncDelay = 2 * time.Minute

const PreviewImageAllowedSizeMaxBytes = 3500000

const UrlImageResizeToWidthPx = 400

/* Variables */

var urlPreviewClient = &http.Client{
	Timeout: 10 * time.Second,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,

		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		IdleConnTimeout:       5 * time.Second,
		MaxConnsPerHost:       0,
		MaxIdleConns:          0,
		MaxIdleConnsPerHost:   0,
	},
}

var urlToPreviewInfoCache = map[string]*UrlPreviewInfoCacheItem{}
var urlToPreviewInfoCacheMutex = sync.Mutex{}

func GetUrlPreviewInfo(url string) (UrlPreviewInfo, error) {
	//if URL is not tracked yet - put it into cache right away (to later utilize cache item lock)
	urlToPreviewInfoCacheMutex.Lock()

	_, urlAlreadyExistsInCache := urlToPreviewInfoCache[url]

	if !urlAlreadyExistsInCache {
		urlToPreviewInfoCache[url] = &UrlPreviewInfoCacheItem{
			urlLock:            sync.Mutex{},
			lastCheckTimestamp: 0,
			lastErrorTimestamp: 0,
			urlPreviewInfoData: UrlPreviewInfo{},
		}
	}

	urlToPreviewInfoCacheMutex.Unlock()

	urlPreviewCacheItem, _ := urlToPreviewInfoCache[url]

	//lock current URL's cache item to prevent concurrent queries to same URL
	urlPreviewCacheItem.urlLock.Lock()
	defer urlPreviewCacheItem.urlLock.Unlock()

	//check if this url preview info was already loaded recently
	if time.Now().UnixNano()-urlPreviewCacheItem.lastCheckTimestamp < UrlToPreviewInfoCacheTTL.Nanoseconds() {

		return urlPreviewCacheItem.urlPreviewInfoData, nil
	}
	//check if this url preview info already errored recently
	if time.Now().UnixNano()-urlPreviewCacheItem.lastErrorTimestamp < UrlToPreviewErrorReCheckDelay.Nanoseconds() {

		return UrlPreviewInfo{}, nil
	}

	//else - load info and put into cache

	urlPreviewCacheItem.lastCheckTimestamp = 0
	urlPreviewCacheItem.lastErrorTimestamp = 0
	urlPreviewCacheItem.urlPreviewInfoData = UrlPreviewInfo{}

	//do request
	r, err := urlPreviewClient.Get(url)

	if err != nil {
		util.LogTrace("Failed to query URL for preview: '%s'. Error: %s", url, err)

		urlPreviewCacheItem.lastErrorTimestamp = time.Now().UnixNano()

		return UrlPreviewInfo{}, err

	} else {
		defer r.Body.Close()

		respContentType, respContentTypeExists := r.Header["Content-Type"]

		/* Check if URL points to image */

		if respContentTypeExists &&
			strings.Contains(strings.Join(respContentType[:], ","), "image") {

			if r.ContentLength > 0 && r.ContentLength <= PreviewImageAllowedSizeMaxBytes {
				//body bytes are actually read from server only at 1st call
				bodyBytes, err := ioutil.ReadAll(r.Body)

				if err != nil {
					util.LogTrace("Failed to read image body for URL '%s'. Error: %s", url, err)

					return UrlPreviewInfo{}, err
				}

				mimeType := http.DetectContentType(bodyBytes)

				if strings.Contains(mimeType, "image") {
					base64DataStr := "data:" + mimeType + ";base64,"

					resizedImgBytes, err := resizeImage(bodyBytes)

					if err != nil {
						util.LogTrace("Failed to resize image for URL '%s'. Error: %s", url, err)

						return UrlPreviewInfo{}, err
					}

					base64DataStr += bytesToBase64(resizedImgBytes)

					urlPreviewCacheItem.lastCheckTimestamp = time.Now().UnixNano()

					urlPreviewCacheItem.urlPreviewInfoData.ImageDataBase64 = &base64DataStr
					urlPreviewCacheItem.urlPreviewInfoData.Url = &url

					return urlPreviewCacheItem.urlPreviewInfoData, nil
				}
			}

			return UrlPreviewInfo{}, err
		}

		/* else - URL must be pointing to web page */

		document, err := goquery.NewDocumentFromReader(r.Body)
		if err != nil {
			util.LogTrace("Failed to load body from URL: '%s'. Error: %s", url, err)

			return UrlPreviewInfo{}, err
		}

		titleStr := ""
		descriptionStr := ""
		descriptionCommonStr := ""
		imageUrlStr := ""

		document.Find("meta").Each(func(index int, element *goquery.Selection) {
			metaTagPropertyAttr, propertyAttrExists := element.Attr("property")

			if propertyAttrExists {
				switch metaTagPropertyAttr {
				case "og:title":
					titleStr, _ = element.Attr("content")
					break

				case "og:description":
					descriptionStr, _ = element.Attr("content")
					break

				case "og:image":
					imageUrlStr, _ = element.Attr("content")
					break
				}

			} else {
				metaTagNameAttr, nameAttrExists := element.Attr("name")

				if nameAttrExists {
					switch metaTagNameAttr {
					case "description":
						descriptionCommonStr, _ = element.Attr("content")
						break
					}
				}
			}
		})

		if titleStr == "" {
			titleStr = document.Find("title").Text()
		}

		if descriptionStr == "" {
			descriptionStr = descriptionCommonStr
		}

		urlPreviewCacheItem.lastCheckTimestamp = time.Now().UnixNano()

		urlPreviewCacheItem.urlPreviewInfoData.Url = &url
		urlPreviewCacheItem.urlPreviewInfoData.Title = &titleStr
		//urlPreviewResponse.Description = &descriptionStr        - field is not used right now
		urlPreviewCacheItem.urlPreviewInfoData.ImageUrl = &imageUrlStr

		return urlPreviewCacheItem.urlPreviewInfoData, nil
	}
}

func StartClearOldCacheItemsFuncPeriodical() {
	ticker := time.NewTicker(ClearOldCacheItemsFuncDelay)

	for {
		select {
		case <-ticker.C:
			clearOldCacheItems()
		}
	}
}

func clearOldCacheItems() {
	var urlsToRemoveFromCacheArr []string

	urlToPreviewInfoCacheMutex.Lock()

	//if item was loaded OK last time (timestamp != 0) but already expired - delete it
	//
	//WORST CONCURRENCY CASE: if we are removing expired cache item right after some user requested that item info (and before it is loaded)
	//- user will just load old cache item instance with some new data and get it returned (though that old item instance will already be removed from cache).
	//If user2 loads in parallel - he will load new cache item instance, which is independent from old one and will be used from now on
	for url, info := range urlToPreviewInfoCache {
		if info.lastCheckTimestamp > 0 &&
			time.Now().UnixNano()-info.lastCheckTimestamp > UrlToPreviewInfoCacheTTL.Nanoseconds() {

			urlsToRemoveFromCacheArr = append(urlsToRemoveFromCacheArr, url)
		}
	}

	for _, url := range urlsToRemoveFromCacheArr {
		delete(urlToPreviewInfoCache, url)
	}

	urlToPreviewInfoCacheMutex.Unlock()
}

func resizeImage(bodyBytes []byte) ([]byte, error) {
	src, err := imaging.Decode(bytes.NewReader(bodyBytes))

	if err != nil {
		return nil, err
	}

	//resize image to width in px, preserving the aspect ratio.
	src = imaging.Resize(src, UrlImageResizeToWidthPx, 0, imaging.Lanczos)

	imgBuffer := bytes.Buffer{}

	err = imaging.Encode(&imgBuffer, src, imaging.PNG)

	if err != nil {
		return nil, err
	}

	return imgBuffer.Bytes(), nil
}

func bytesToBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
