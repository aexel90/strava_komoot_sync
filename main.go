package main

import (
	"flag"
	"log"
	"math"
	"strings"
	"time"

	stravaLib "github.com/aexel90/go.strava"

	"github.com/aexel90/strava_komoot_sync/constants"
	"github.com/aexel90/strava_komoot_sync/komoot"
	"github.com/aexel90/strava_komoot_sync/strava"
)

var (
	komootEmail             = flag.String("komoot_email", "", "Komoot Email")
	komootPassword          = flag.String("komoot_pw", "", "Komoot Password")
	komootUserId            = flag.String("komoot_userid", "", "Komoot User ID")
	stravaClientId          = flag.Int("strava_clientid", 0, "Strava Client ID")
	stravaClientSecret      = flag.String("strava_clientsecret", "", "Strava Client Secret")
	stravaAthleteId         = flag.Int64("strava_athleteid", 0, "Strava Athlete ID")
	syncAll                 = flag.Bool("sync_all", false, "Sync all activities")
	stravaVirtualRideGearId = flag.String("strava_virtualRide_gearid", "", "Strava Virtual Ride GearID")
)

func main() {
	flag.Parse()

	komootService := komoot.NewKomootService(*komootEmail, *komootPassword, *komootUserId)
	stravaService := strava.NewStravaService(*stravaClientId, *stravaClientSecret, *stravaAthleteId)

	// sync all and quit
	if *syncAll {
		sync(stravaService, komootService, *syncAll, *stravaVirtualRideGearId)
		return
	}

	// sync endless loop
	for {
		sync(stravaService, komootService, false, *stravaVirtualRideGearId)
		time.Sleep(5 * time.Minute)
	}
}

func sync(stravaService *strava.StravaService, komootService *komoot.KomootService, syncAll bool, stravaVirtualRideGearId string) {

	log.Print("******************************* NEW SYNC LOOP ******************************")

	stravaActivities, err := stravaService.GetActivities(syncAll)
	if err != nil {
		log.Printf("STRAVA - GetActivities ERROR: %s", err)
		return
	}

	komootActivities, err := komootService.GetActivities(syncAll)
	if err != nil {
		log.Printf("KOMOOT - GetActivities ERROR: %s", err)
		return
	}

	for _, stravaActivity := range stravaActivities {

		log.Print("****************************************************************************")
		log.Printf("STRAVA: Id: '%d'\tDate: '%s' Name: '%s' Distance: '%f' Private: %t", stravaActivity.Id, stravaActivity.StartDate.Format(constants.TimeFormat), stravaActivity.Name, stravaActivity.Distance, stravaActivity.Private)

		// VIRTUAL RIDEs
		if stravaVirtualRideGearId != "" && stravaActivity.Type == stravaLib.ActivityTypes.VirtualRide {

			log.Printf("STRAVA: Id: '%d'\tType: '%s' GearId: '%s'", stravaActivity.Id, stravaActivity.Type, stravaActivity.GearId)

			var updateRequired bool

			if stravaActivity.GearId != stravaVirtualRideGearId {
				stravaActivity.GearId = stravaVirtualRideGearId
				updateRequired = true
			}

			// set private --> public - still not supported :(
			/*if stravaActivity.Private {
				stravaActivity.Private = false
				updateRequired = true
			} */

			if updateRequired {
				err := stravaService.UpdateActivity(stravaActivity)
				if err != nil {
					log.Printf("STRAVA: Activity Update ERROR: %s", err)
				}
			}

			// ALL OTHER RIDES
		} else {

			komootActivity := getActivityMatch(stravaActivity.StartDate, stravaActivity.Distance, komootActivities)
			if komootActivity == nil {
				log.Print("KOMOOT: no activity found")
				continue
			}

			log.Printf("KOMOOT: Id: '%d'\tDate: '%s' Name: '%s' Distance: '%f' Private: %t", komootActivity.ID, komootActivity.Date.Format(constants.TimeFormat), komootActivity.Name, komootActivity.Distance, komootActivity.Private)

			//check komoot name
			var newKomootName string
			if strings.TrimSpace(stravaActivity.Name) != strings.TrimSpace(komootActivity.Name) {
				newKomootName = stravaActivity.Name
			}

			// check komoot visibility
			var public bool
			if !stravaActivity.Private && komootActivity.Private {
				public = true
			}

			// update komoot
			if newKomootName == "" && !public {
				continue
			} else {
				err := komootService.UpdateActivity(komootActivity, newKomootName, public)
				if err != nil {
					log.Printf("KOMOOT - Update Activity ERROR: %s", err)
				}
			}
		}
	}
}

func getActivityMatch(stravaDate time.Time, stravaDistance float64, komootActivities *[]komoot.Activity) *komoot.Activity {

	for _, komootActivity := range *komootActivities {
		// distance tolerance 3km - date tolerance 1 hour
		if math.Abs(stravaDate.Sub(komootActivity.Date).Hours()) < 1 && math.Abs(stravaDistance-komootActivity.Distance) < 3000 {
			return &komootActivity
		}
	}
	return nil
}
