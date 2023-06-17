package yt_transcript

import (
	"testing"
)

func TestFetchTranscript(t *testing.T) {
	videoId := "wqbeCG-Y3Xk"
	transcripts, title, err := FetchTranscript(videoId, "en", "US")
	if err != nil {
		t.Errorf("FetchTranscript failed with error: %v", err)
	}

	if title == "" {
		t.Error("FetchTranscript returned empty title")
	}

	if len(transcripts) == 0 {
		t.Error("FetchTranscript returned empty transcripts")
	}

	for _, transcript := range transcripts {
		if transcript.Text == "" {
			t.Error("Transcript has empty text")
		}
		if transcript.Duration <= 0 {
			t.Errorf("Transcript has invalid duration: %d", transcript.Duration)
		}
		if transcript.Offset < 0 {
			t.Errorf("Transcript has invalid offset: %d", transcript.Offset)
		}
	}

}

func TestFetchTranscript2(t *testing.T) {
	videoId := "Rt78MqJDozY"
	transcripts, title, err := FetchTranscript(videoId, "en", "US")
	if err != nil {
		t.Errorf("FetchTranscript failed with error: %v", err)
	}

	if title == "" {
		t.Error("FetchTranscript returned empty title")
	}

	if len(transcripts) == 0 {
		t.Error("FetchTranscript returned empty transcripts")
	}

	for _, transcript := range transcripts {
		if transcript.Text == "" {
			t.Error("Transcript has empty text")
		}
		if transcript.Duration <= 0 {
			t.Errorf("Transcript has invalid duration: %d", transcript.Duration)
		}
		if transcript.Offset < 0 {
			t.Errorf("Transcript has invalid offset: %d", transcript.Offset)
		}
	}

}

func TestGetVideoTitle(t *testing.T) {
	videoId := "dQw4w9WgXcQ"
	title, err := GetVideoTitle(videoId)
	if err != nil {
		t.Errorf("getVideoTitle(%s) returned an error: %s", videoId, err.Error())
	} else if title != "Rick Astley - Never Gonna Give You Up (Official Music Video)" {
		t.Errorf("getVideoTitle(%s) returned the wrong title: %s", videoId, title)
	}
}

func TestGetVideoID(t *testing.T) {
	url := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	videoId, err := GetVideoID(url)
	if err != nil {
		t.Errorf("GetVideoID(%s) returned an error: %s", url, err.Error())
	} else if videoId != "dQw4w9WgXcQ" {
		t.Errorf("GetVideoID(%s) returned the wrong video ID: %s", url, videoId)
	}

	url = "https://youtu.be/dQw4w9WgXcQ"
	videoId, err = GetVideoID(url)
	if err != nil {
		t.Errorf("GetVideoID(%s) returned an error: %s", url, err.Error())
	} else if videoId != "dQw4w9WgXcQ" {
		t.Errorf("GetVideoID(%s) returned the wrong video ID: %s", url, videoId)
	}

	url = "https://www.invalid.com/watch?v=dQw4w9WgXcQ"
	_, err = GetVideoID(url)
	if err == nil {
		t.Errorf("GetVideoID(%s) did not return an error for an invalid URL", url)
	}
}
