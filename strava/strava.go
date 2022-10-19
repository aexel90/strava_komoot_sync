package strava

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	strava "github.com/aexel90/go.strava"
	"github.com/aexel90/strava_komoot_sync/constants"
)

type StravaService struct {
	ClientID     int
	ClientSecret string
	AthleteId    int64
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
	ExpiresIn    int
}

const port = 8080
const tokenFile = "token.json"

var c = make(chan *strava.AuthorizationResponse)
var authenticator *strava.OAuthAuthenticator

func NewStravaService(clientID int, clientSecret string, athleteId int64) *StravaService {
	strava.ClientId = clientID
	strava.ClientSecret = clientSecret

	s := StravaService{}
	s.ClientID = clientID
	s.ClientSecret = clientSecret
	s.AthleteId = athleteId

	authenticator = &strava.OAuthAuthenticator{
		CallbackURL:            fmt.Sprintf("http://localhost:%d/exchange_token", port),
		RequestClientGenerator: nil,
	}

	return &s
}

func (s *StravaService) handleAuthorizationResponse(authResp *strava.AuthorizationResponse) {

	s.AccessToken = authResp.AccessToken
	s.RefreshToken = authResp.RefreshToken
	s.ExpiresAt = authResp.ExpiresAt
	s.ExpiresIn = authResp.ExpiresIn

	s.printTokenDetails()
	s.storeTokensToFile()
}

func (s *StravaService) printTokenDetails() {

	log.Printf("AccessToken:\t\t%s", s.AccessToken)
	log.Printf("RefreshToken:\t\t%s", s.RefreshToken)
	log.Printf("AccessToken expires at:\t%s", time.Unix(s.ExpiresAt, 0).Format(constants.TimeFormat))
	log.Printf("AccessToken expires in:\t%ds", s.ExpiresIn)

}

func (s *StravaService) GetActivities(syncAll bool) ([]*strava.ActivitySummary, error) {
	client, err := s.getStravaClient()
	if err != nil {
		return nil, err
	}
	/* Requests that return multiple items will be paginated to 30 items by default.
	The page parameter can be used to specify further pages or offsets. The per_page may also be used for custom page sizes up to 200.
	Note that in certain cases, the number of items returned in the response may be lower than the requested page size, even when that page is not the last.
	If you need to fully go through the full set of results, prefer iterating until an empty page is returned. */
	if syncAll {
		var allResults []*strava.ActivitySummary
		i := 1
		for {
			pageResult, err := strava.NewAthletesService(client).ListActivities(s.AthleteId).Page(i).Do()

			if err != nil {
				return nil, err
			}
			allResults = append(allResults, pageResult...)
			log.Printf("StravaRequest Page%d:\t%d results (sum: %d)", i, len(pageResult), len(allResults))

			if len(pageResult) == 0 {
				break
			}
			i++
		}
		return allResults, nil

	} else {
		return strava.NewAthletesService(client).ListActivities(s.AthleteId).Page(1).Do()
	}
}

func (s *StravaService) GetActivityDetails(activityId int64) (*strava.ActivityDetailed, error) {
	client, err := s.getStravaClient()
	if err != nil {
		return nil, err
	}
	return strava.NewActivitiesService(client).Get(activityId).Do()
}

func (s *StravaService) getStravaClient() (*strava.Client, error) {

	if s.AccessToken == "" {
		s.loadTokensFromFile()
	}

	if s.AccessToken == "" {
		go s.getAcessToken()
		s.handleAuthorizationResponse(<-c)
	}

	if s.ExpiresAt < time.Now().Unix() {
		err := s.refreshAccessToken()
		if err != nil {
			return nil, err
		}
	}
	return strava.NewClient(s.AccessToken), nil
}

func (s *StravaService) refreshAccessToken() error {

	log.Print("Refreshing AccessToken ...")

	newRefreshToken, err := authenticator.RefreshAuthorize(s.RefreshToken, nil)
	if err != nil {
		return err
	}

	s.handleAuthorizationResponse(newRefreshToken)
	return nil
}

func (s *StravaService) getAcessToken() {

	path, err := authenticator.CallbackPath()
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc(path, authenticator.HandlerFunc(oAuthSuccess, oAuthFailure))

	fmt.Printf("Accept Strava Access: %s\n", authenticator.AuthorizationURL("state1", strava.Permissions.ActivityReadAll, true))
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		panic(err)
	}
}

func (s *StravaService) loadTokensFromFile() {
	jsonData, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		log.Printf("File '%s' does not exist.", tokenFile)
		return
	}

	err = json.Unmarshal(jsonData, s)
	if err != nil {
		log.Printf("Error parsing tokenFile: %v", err)
		return
	}

	log.Print("Token read from tokenFile.")
	s.printTokenDetails()
}

func (s *StravaService) storeTokensToFile() {
	jsonData, err := json.MarshalIndent(s, "", " ")
	if err != nil {
		log.Printf("Error creating json data to store Token into file: %v", jsonData)
		return
	}

	err = ioutil.WriteFile(tokenFile, jsonData, 0644)
	if err != nil {
		log.Printf("Error writing tokenFile: %v", err)
		return
	}

	log.Print("Token stored in tokenFile.")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `<a href="%s">`, authenticator.AuthorizationURL("state1", strava.Permissions.ActivityReadAll, true))
	fmt.Fprint(w, `<img src="http://strava.github.io/api/images/ConnectWithStrava.png" />`)
	fmt.Fprint(w, `</a>`)
}

func oAuthSuccess(auth *strava.AuthorizationResponse, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Access Token: %s\n", auth.AccessToken)
	fmt.Fprintf(w, "Refresh Token: %s\n", auth.RefreshToken)
	fmt.Fprintf(w, "Access Token expires at: %s\n", time.Unix(auth.ExpiresAt, 0))
	fmt.Fprintf(w, "Access Token expires in (sec): %d\n\n", auth.ExpiresIn)

	fmt.Fprintf(w, "The Authenticated Athlete (you):\n")
	content, _ := json.MarshalIndent(auth.Athlete, "", " ")
	fmt.Fprint(w, string(content))

	c <- auth
}

func oAuthFailure(err error, w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Authorization Failure:\n")

	if err == strava.OAuthAuthorizationDeniedErr {
		fmt.Fprint(w, "The user clicked the 'Do not Authorize' button on the previous page.\n")
		fmt.Fprint(w, "This is the main error your application should handle.")
	} else if err == strava.OAuthInvalidCredentialsErr {
		fmt.Fprint(w, "You provided an incorrect client_id or client_secret.\nDid you remember to set them at the begininng of this file?")
	} else if err == strava.OAuthInvalidCodeErr {
		fmt.Fprint(w, "The temporary token was not recognized, this shouldn't happen normally")
	} else if err == strava.OAuthServerErr {
		fmt.Fprint(w, "There was some sort of server error, try again to see if the problem continues")
	} else {
		fmt.Fprint(w, err)
	}

	c <- &strava.AuthorizationResponse{}
}
