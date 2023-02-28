package komoot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"
)

type KomootService struct {
	email            string
	password         string
	userid           string
	clientLoginAlive bool
	httpClient       *http.Client
}

const komootSignInURL = "https://account.komoot.com/v1/signin"
const komootTransferURL = "https://account.komoot.com/actions/transfer?type=signin"
const komootApiURL = "https://www.komoot.de/api/v007"
const statusFriends = "friends"

const ActivityDateFormat1 = "2006-01-02T15:04:05.000Z"
const ActivityDateFormat2 = "2006-01-02T15:04:05.000-07:00"

func NewKomootService(email string, password string, userid string) *KomootService {
	return &KomootService{email, password, userid, false, nil}
}

func (k *KomootService) getHttpClient() (*http.Client, error) {

	if k.httpClient == nil {

		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, fmt.Errorf("got error while creating cookie jar: %s", err)
		}
		k.httpClient = &http.Client{Jar: jar}

	}
	if !k.clientLoginAlive {

		//SignIn
		resp, err := k.httpClient.PostForm(komootSignInURL, url.Values{
			"password": {k.password},
			"email":    {k.email},
		})
		if err != nil {
			return nil, err
		}
		io.Copy(io.Discard, resp.Body)
		defer resp.Body.Close()

		status := fmt.Sprintf("%s: %s", komootSignInURL, resp.Status)
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New(status)
		}
		log.Debug(status)

		//Tranfer
		resp, err = k.httpClient.Get(komootTransferURL)
		if err != nil {
			return nil, err
		}
		io.Copy(io.Discard, resp.Body)
		defer resp.Body.Close()

		status = fmt.Sprintf("%s: %s", komootTransferURL, resp.Status)
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New(status)
		}
		log.Debug(status)

		k.clientLoginAlive = true
	}
	return k.httpClient, nil
}

func (k *KomootService) GetActivities(syncAll bool) (data *[]Activity, err error) {

	if syncAll {
		var allResults []Activity
		i := 0
		for {
			pageResult, err := k.requestActivities(i)

			if err != nil {
				return nil, err
			}
			allResults = append(allResults, *pageResult...)
			log.Debugf("KomootRequest Page%d:\t%d results (sum: %d)", i, len(*pageResult), len(allResults))

			if len(*pageResult) == 0 {
				break
			}
			i++
		}
		return &allResults, nil
	} else {
		return k.requestActivities(0)
	}
}

func (k *KomootService) requestActivities(page int) (data *[]Activity, err error) {

	httpClient, err := k.getHttpClient()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/users/%s/tours/?type=tour_recorded&limit=100&page=%d", komootApiURL, k.userid, page)
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	status := fmt.Sprintf("%s: %s", url, resp.Status)
	if resp.StatusCode != http.StatusOK {
		k.clientLoginAlive = false
		return nil, errors.New(status)
	}
	log.Debug(status)

	var toursResponse ToursResponse
	err = json.Unmarshal(body, &toursResponse)
	if err != nil {
		return nil, err
	}

	for i := range toursResponse.Embedded.Tours {
		setAccess(toursResponse, i)
		err := setDate(toursResponse, i)
		if err != nil {
			return nil, err
		}
	}
	return &toursResponse.Embedded.Tours, nil
}

func setAccess(toursResponse ToursResponse, i int) {
	if toursResponse.Embedded.Tours[i].Status != statusFriends {
		toursResponse.Embedded.Tours[i].Private = true
	}
}

func setDate(toursResponse ToursResponse, i int) error {
	// seen TimeFormats
	// "2021-06-27T08:31:40.000Z"  		=>   2006-01-02T15:04:05.000Z
	// "2021-06-27T10:30:43.716+02:00"	=>   2021-06-27T10:30:43.716+02:00

	// 1.
	komootActivityDate, err1 := time.Parse(ActivityDateFormat1, toursResponse.Embedded.Tours[i].DateString)
	if err1 == nil {
		toursResponse.Embedded.Tours[i].Date = komootActivityDate
		return nil
	}

	// 2.
	komootActivityDate, err2 := time.Parse(ActivityDateFormat2, toursResponse.Embedded.Tours[i].DateString)
	if err2 == nil {
		toursResponse.Embedded.Tours[i].Date = komootActivityDate
		return nil
	}
	return errors.New(err1.Error() + "\n" + err2.Error())
}

func (k *KomootService) UpdateActivity(komootActivity *Activity, name string, public bool) error {

	httpClient, err := k.getHttpClient()
	if err != nil {
		return err
	}

	data := make(map[string]interface{}, 2)
	if name != "" {
		log.Debugf("Updating KomootActivity '%d' - Name: '%s' --> '%s'", komootActivity.ID, komootActivity.Name, name)
		data["name"] = name
	}
	if public {
		log.Debugf("Updating KomootActivity '%d' - Visibility: '%s' --> '%s'", komootActivity.ID, komootActivity.Status, statusFriends)
		data["status"] = statusFriends
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/tours/%v", komootApiURL, komootActivity.ID)
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	status := fmt.Sprintf("%s: %s", url, resp.Status)
	if resp.StatusCode != http.StatusOK {
		return errors.New(status + string(body))
	}
	log.Debug(status)

	return nil
}
