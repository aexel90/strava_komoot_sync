# strava_komoot_sync

Activity Synchronization between Strava and Komoot

## Synchronization direction
Strava --> Komoot

## What is synced
- Name of the activity

## Execution

    $GOPATH/bin/strava_komoot_sync -h

    Usage of ./strava_komoot_sync:
            -komoot_email string
                    Komoot Email
            -komoot_pw string
                    Komoot Password
            -komoot_userid string
                    Komoot User ID
            -strava_athleteid int
                    Strava Athlete ID
            -strava_clientid int
                    Strava Client ID
            -strava_clientsecret string
                    Strava Client Secret
            -strava_code string
                    Strava Code
            -sync_all
                    Sync all activities

## Flag -sync_all
- true:  all activities will be synched once and program terminates
- false: the last 30 Strava activities will be synched each 5 minutes

## ToDos
- store AccessToken across Sessions
- sync pics
- sync visibility
- docker config