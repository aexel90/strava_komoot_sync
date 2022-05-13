package komoot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

type KomootService struct {
	email      string
	password   string
	userid     string
	httpClient *http.Client
}

const komootSignInURL = "https://account.komoot.com/v1/signin"
const komootTransferURL = "https://account.komoot.com/actions/transfer?type=signin"
const komootApiURL = "https://www.komoot.de/api/v007"
const statusFriends = "friends"

const ActivityDateFormat1 = "2006-01-02T15:04:05.000Z"
const ActivityDateFormat2 = "2006-01-02T15:04:05.000-07:00"

func NewKomootService(email string, password string, userid string) *KomootService {
	return &KomootService{email, password, userid, nil}
}

func (k *KomootService) getHttpClient() (*http.Client, error) {

	if k.httpClient == nil {
		// Cookie Setup
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, fmt.Errorf("got error while creating cookie jar: %s", err)
		}
		client := http.Client{Jar: jar}
		log.Println("Cookie initialized")

		//SignIn
		resp, err := client.PostForm(komootSignInURL, url.Values{
			"password": {k.password},
			"email":    {k.email},
		})
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		status := fmt.Sprintf("%s: %s", komootSignInURL, resp.Status)
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New(status)
		}
		log.Print(status)

		//Tranfer
		resp, err = client.Get(komootTransferURL)
		if err != nil {
			return nil, err
		}
		status = fmt.Sprintf("%s: %s", komootTransferURL, resp.Status)
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New(status)
		}
		log.Print(status)

		k.httpClient = &client
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
			log.Printf("KomootRequest Page%d:\t%d results (sum: %d)", i, len(*pageResult), len(allResults))

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

	status := fmt.Sprintf("%s: %s", url, resp.Status)
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(status)
	}
	log.Print(status)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var toursResponse ToursResponse
	err = json.Unmarshal(body, &toursResponse)
	if err != nil {
		return nil, err
	}

	// seen TimeFormats
	// "2021-06-27T08:31:40.000Z"  		=>   2006-01-02T15:04:05.000Z
	// "2021-06-27T10:30:43.716+02:00"	=>   2021-06-27T10:30:43.716+02:00
	for i := range toursResponse.Embedded.Tours {

		if toursResponse.Embedded.Tours[i].Status != statusFriends {
			toursResponse.Embedded.Tours[i].Private = true
		}

		// 1.
		komootActivityDate, err1 := time.Parse(ActivityDateFormat1, toursResponse.Embedded.Tours[i].DateString)
		if err1 == nil {
			toursResponse.Embedded.Tours[i].Date = komootActivityDate
			continue
		}

		// 2.
		komootActivityDate, err2 := time.Parse(ActivityDateFormat2, toursResponse.Embedded.Tours[i].DateString)
		if err2 == nil {
			toursResponse.Embedded.Tours[i].Date = komootActivityDate
			continue
		}
		return nil, errors.New(err1.Error() + "\n" + err2.Error())
	}
	return &toursResponse.Embedded.Tours, nil
}

func (k *KomootService) UpdateActivity(komootActivity *Activity, name string, public bool) error {

	httpClient, err := k.getHttpClient()
	if err != nil {
		return err
	}

	data := make(map[string]interface{}, 2)
	if name != "" {
		log.Printf("Updating KomootActivity '%d' - Name: '%s' --> '%s'", komootActivity.Id, komootActivity.Name, name)
		data["name"] = name
	}
	if public {
		log.Printf("Updating KomootActivity '%d' - Visibility: '%s' --> '%s'", komootActivity.Id, komootActivity.Status, statusFriends)
		data["status"] = statusFriends
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/tours/%v", komootApiURL, komootActivity.Id)
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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	status := fmt.Sprintf("%s: %s", url, resp.Status)
	if resp.StatusCode != http.StatusOK {
		return errors.New(status + string(body))
	}
	log.Print(status)

	return nil
}
