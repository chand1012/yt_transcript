# YT Transcript

This is a Go package for retrieving the transcript of a YouTube video. It provides a simple interface for fetching the transcript of a video by its ID.

## Installation

You can install the package using Go modules:

```sh

go get github.com/chand1012/yt_transcript
```

## Usage

Here is an example of how to use the package to fetch the transcript of a video:

```go

package main

import (
    "fmt"
    "github.com/chand1012/yt_transcript"
)

func main() {
    videoURL := "https://youtu.be/wqbeCG-Y3Xk"
    videoId, err := yt_transcript.GetVideoId(videoURL)
    if err != nil {
        panic(err)
    }
	  transcripts, title, err := yt_transcript.FetchTranscript(videoId, "en", "US")
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Printf("Title: %s\n", title)
        for _, t := range transcript {
            fmt.Printf("[%d:%d] %s\n", t.Offset/1000, (t.Offset+t.Duration)/1000, t.Text)
        }
    }
}
```

The FetchTranscript method takes a YouTube video ID and a TranscriptConfig object as arguments. The TranscriptConfig object specifies the language and country of the transcript.

The method returns a slice of TranscriptResponse objects, which contain the text of each caption, along with the start time and duration of the caption. It also returns the title of the video.

## License

This package is licensed under the MIT License. See the LICENSE file for details.
