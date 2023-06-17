package yt_transcript

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"math/big"
	"net/http"
	"regexp"
	"strconv"
	
)

var reYoutube = regexp.MustCompile(`^.*(?:(?:youtu\.be\/|v\/|vi\/|u\/\w\/|embed\/)|(?:(?:watch)?\?v(?:i)?=|\&v(?:i)?=))([^#\&\?]*).*`)

type YoutubeTranscriptError struct {
	Message string
}

func (e *YoutubeTranscriptError) Error() string {
	return fmt.Sprintf("[YoutubeTranscript] 🚨 %s", e.Message)
}

type ytConfig struct {
	Lang    string
	Country string
}

type TranscriptResponse struct {
	Text     string
	Duration int
	Offset   int
}

type ytTranscript struct{}

// FetchTranscript fetches the transcript for a given video ID
func FetchTranscript(videoID, lang, country string) ([]TranscriptResponse, string, error) {
	yt := ytTranscript{}
	config := &ytConfig{Lang: lang, Country: country}
	return yt.fetchTranscript(videoID, config)
}
func (yt *ytTranscript) fetchTranscript(videoId string, config *ytConfig) ([]TranscriptResponse, string, error) {
	identifier, err := GetVideoID(videoId)
	
	if err != nil {
		return nil, "", &YoutubeTranscriptError{Message: err.Error()}
	}

	resp, err := http.Get(fmt.Sprintf("https://www.youtube.com/watch?v=%s", identifier))
	if err != nil {
		return nil, "", &YoutubeTranscriptError{Message: err.Error()}
	}
	defer resp.Body.Close()
	videoPageBody, _ := io.ReadAll(resp.Body)

	innerTubeApiKey := regexp.MustCompile(`"INNERTUBE_API_KEY":"(.*?)"`).FindStringSubmatch(string(videoPageBody))[1]
	reTitle := regexp.MustCompile(`(?i)<title>.*?([^<>]*)</title>`)
	titleMatch := reTitle.FindStringSubmatch(string(videoPageBody))
	title := ""
	
	if len(titleMatch) > 1 {
		title = titleMatch[1]
	}	

	// remove " - YouTube" from title
	title = html.UnescapeString(regexp.MustCompile(`(?i)\s-\sYouTube$`).ReplaceAllString(title, ""))
	
	if len(innerTubeApiKey) > 0 {
		client := &http.Client{}
		reqBody, _ := json.Marshal(yt.generateReq(string(videoPageBody), config))
		req, _ := http.NewRequest("POST", fmt.Sprintf("https://www.youtube.com/youtubei/v1/get_transcript?key=%s", innerTubeApiKey), bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		res, err := client.Do(req)
		if err != nil {
			return nil, "", &YoutubeTranscriptError{Message: err.Error()}
		}
		defer res.Body.Close()

		var body map[string]interface{}
		json.NewDecoder(res.Body).Decode(&body)
		if _, ok := body["responseContext"]; ok {
			if _, ok := body["actions"]; !ok {
				return nil, "", &YoutubeTranscriptError{Message: "Transcript is disabled on this video"}
			}

			transcriptsRaw := body["actions"].([]interface{})[0].(map[string]interface{})["updateEngagementPanelAction"].(map[string]interface{})["content"].(map[string]interface{})["transcriptRenderer"].(map[string]interface{})["body"].(map[string]interface{})["transcriptBodyRenderer"].(map[string]interface{})["cueGroups"].([]interface{})
			var transcripts []TranscriptResponse
			for _, cue := range transcriptsRaw {
				cueGroupRenderer := cue.(map[string]interface{})["transcriptCueGroupRenderer"].(map[string]interface{})["cues"].([]interface{})[0].(map[string]interface{})["transcriptCueRenderer"].(map[string]interface{})
				duration, _ := strconv.Atoi(cueGroupRenderer["durationMs"].(string))
				offset, _ := strconv.Atoi(cueGroupRenderer["startOffsetMs"].(string))
				
				// if there is a blank part of the transcript, skip it
				if(cueGroupRenderer["cue"].(map[string]interface{})["simpleText"] == nil){
					continue
				}

				transcripts = append(transcripts, TranscriptResponse{
					Text:     cueGroupRenderer["cue"].(map[string]interface{})["simpleText"].(string),
					Duration: duration,
					Offset:   offset,
				})
			}

			return transcripts, title, nil
		}
	}

	return nil, "", &YoutubeTranscriptError{Message: "Failed to fetch transcript"}
}

func (yt *ytTranscript) generateReq(page string, config *ytConfig) map[string]interface{} {

	paramsMatch := regexp.MustCompile(`"serializedShareEntity":"(.*?)"`).FindStringSubmatch(page)
	visitorDataMatch := regexp.MustCompile(`"VISITOR_DATA":"(.*?)"`).FindStringSubmatch(page)
	// sessionIdMatch := regexp.MustCompile(`"sessionId":"(.*?)"`).FindStringSubmatch(page)
	clickTrackingParamsMatch := regexp.MustCompile(`"clickTrackingParams":"(.*?)"`).FindStringSubmatch(page)

	// if len(paramsMatch) < 2 || len(visitorDataMatch) < 2 || len(sessionIdMatch) < 2 || len(clickTrackingParamsMatch) < 2 {
	if len(paramsMatch) < 2 || len(visitorDataMatch) < 2 || len(clickTrackingParamsMatch) < 2 {
		panic(&YoutubeTranscriptError{Message: "Failed to extract required data from the page"})
	}

	params := paramsMatch[1]
	visitorData := visitorDataMatch[1]
	// sessionId := sessionIdMatch[1]
	clickTrackingParams := clickTrackingParamsMatch[1]

	return map[string]interface{}{
		"context": map[string]interface{}{
			"client": map[string]interface{}{
				"hl":                 config.Lang,
				"gl":                 config.Country,
				"visitorData":        visitorData,
				"userAgent":          "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.83 Safari/537.36,gzip(gfe)",
				"clientName":         "WEB",
				"clientVersion":      "2.20200925.01.00",
				"osName":             "Macintosh",
				"osVersion":          "10_15_4",
				"browserName":        "Chrome",
				"browserVersion":     "85.0f.4183.83",
				"screenWidthPoints":  1440,
				"screenHeightPoints": 770,
				"screenPixelDensity": 2,
				"utcOffsetMinutes":   120,
				"userInterfaceTheme": "USER_INTERFACE_THEME_LIGHT",
				"connectionType":     "CONN_CELLULAR_3G",
			},
			"request": map[string]interface{}{
				// "sessionId":               sessionId,
				"internalExperimentFlags": []interface{}{},
				"consistencyTokenJars":    []interface{}{},
			},
			"user":              map[string]interface{}{},
			"clientScreenNonce": yt.generateNonce(),
			"clickTracking": map[string]interface{}{
				"clickTrackingParams": clickTrackingParams,
			},
		},
		"params": params,
	}
}

func (yt *ytTranscript) generateNonce() string {
	alphabet := "ABCDEFGHIJKLMOPQRSTUVWXYZabcdefghjijklmnopqrstuvwxyz0123456789"
	nonce := make([]byte, 16)
	rand.Read(nonce)
	for i := 0; i < len(nonce); i++ {
		randomIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		nonce[i] = alphabet[randomIdx.Int64()]
	}
	return base64.RawURLEncoding.EncodeToString(nonce)
}

// Gets the ID of a video given a URL
func GetVideoID(url string) (string, error) {
	// make sure it only contains valid video ID characters
	// valid characters are 0-9, A-Z, a-z, -, and _
	re := regexp.MustCompile(`^[0-9A-Za-z_-]{11}$`)
	if len(url) == 11 {
		if re.MatchString(url) {
			return url, nil
		}
		return "", &YoutubeTranscriptError{Message: "Invalid Youtube video ID."}
	}
	// if the string does not start with youtube.com, youtu.be, or youtube-nocookie.com
	// then it is not a valid youtube video url
	reURL := regexp.MustCompile(`(?i)(?:youtube(?:-nocookie)?\.com/(?:[^/]+/.+/|(?:v|e(?:mbed)?)/|.*[?&]v=)|youtu\.be/)([^"&?/ ]{11})`)
	matchURL := reURL.FindStringSubmatch(url)
	if len(matchURL) == 0 {
		return "", &YoutubeTranscriptError{Message: "Invalid Youtube video URL."}
	}
	matchId := reYoutube.FindStringSubmatch(url)
	if len(matchId) > 0 {
		if re.MatchString(matchId[1]) {
			return matchId[1], nil
		}
		return "", &YoutubeTranscriptError{Message: "Invalid Youtube video ID."}
	}
	return "", &YoutubeTranscriptError{Message: "Impossible to retrieve Youtube video ID."}
}

// Gets the title of a video given its ID
func GetVideoTitle(videoId string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoId))
	if err != nil {
		return "", &YoutubeTranscriptError{Message: err.Error()}
	}
	defer resp.Body.Close()
	videoPageBody, _ := io.ReadAll(resp.Body)

	reTitle := regexp.MustCompile(`(?i)<title>.*?([^<>]*)</title>`)
	titleMatch := reTitle.FindStringSubmatch(string(videoPageBody))
	title := ""
	if len(titleMatch) > 1 {
		title = titleMatch[1]
	}

	// remove " - YouTube" from title
	title = html.UnescapeString(regexp.MustCompile(`(?i)\s-\sYouTube$`).ReplaceAllString(title, ""))
	if title == "" {
		return "", &YoutubeTranscriptError{Message: "Failed to fetch video title"}
	}
	return title, nil
}
