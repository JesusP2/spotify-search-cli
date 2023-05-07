package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type SpotifyTokenRequest struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Expires     int    `json:"expires"`
}

type ExternalUrls struct {
	Spotify string `json:"spotify"`
}

type ArtistBody struct {
	Id           string       `json:"id"`
	Genres       []string     `json:"genres"`
	Name         string       `json:"name"`
	Popularity   int          `json:"popularity"`
	ExternalUrls ExternalUrls `json:"external_urls"`
}

type AlbumBody struct {
	Id           string       `json:"id"`
	Name         string       `json:"name"`
	ReleaseDate  string       `json:"release_date"`
	TotalTracks  int          `json:"total_tracks"`
	AlbumType    string       `json:"album_type"`
	ExtrenalUrls ExternalUrls `json:"external_urls"`
}

type SearchBody struct {
	Artists struct {
		Items []ArtistBody `json:"items"`
	}
}

const (
	Album    = "album"
	Artist   = "artist"
	Playlist = "playlist"
	Track    = "track"
)

func RequestSpotifyToken() SpotifyTokenRequest {
	godotenv.Load(".env")
	client_id := os.Getenv("CLIENT_ID")
	client_secret := os.Getenv("CLIENT_SECRET")

	data := url.Values{
		"grant_type":    []string{"client_credentials"},
		"client_id":     []string{client_id},
		"client_secret": []string{client_secret},
	}
	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal("Error creating request:", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c := http.Client{Timeout: 10 * time.Second}
	res, err := c.Do(req)
	if err != nil {
		log.Fatal("Error making request:", err)
	}
	defer res.Body.Close()
	var body []byte
	buf := make([]byte, 4)
	for {
		n, err := res.Body.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("Error reading data:", err)
		}
		body = append(body, buf[:n]...)
	}

	var myBody SpotifyTokenRequest

	err = json.Unmarshal(body, &myBody)
	if err != nil {
		log.Fatal("Error parsing body:", err)
	}

	return myBody
}

func Search(q string, searchType string, token string, c *http.Client) SearchBody {
	queryParams := url.Values{
		"query":  []string{q},
		"type":   []string{searchType},
		"market": []string{"US"},
		"limit":  []string{"10"},
		"offset": []string{"0"},
	}
	url := &url.URL{
		Scheme:   "https",
		Host:     "api.spotify.com",
		Path:     "v1/search",
		RawQuery: queryParams.Encode(),
	}
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		log.Fatal("Error creating request:", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	res, err := c.Do(req)
	if err != nil {
		log.Fatal("Error making request:", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	var ArtistBody SearchBody
	err = json.Unmarshal(body, &ArtistBody)
	return ArtistBody
}

func GetArtistData(id string, token string, c *http.Client) {
	path := fmt.Sprintf("%s/%s", "v1/artists", id)
	url := &url.URL{
		Scheme: "https",
		Host:   "api.spotify.com",
		Path:   path,
	}
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		log.Fatal("Error creating request:", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	res, err := c.Do(req)
	if err != nil {
		log.Fatal("Error making request:", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	fmt.Println(string(body))
}
