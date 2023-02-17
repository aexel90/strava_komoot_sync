# StravaKomootSync

Activity Synchronization between Strava and Komoot.
Synchronization direction: Strava --> Komoot

## What is synced
- Name of the activity
- Visibility (private, public)

## Binary

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
        -sync_all
                Sync all activities

### Flag -sync_all
- true:  all activities will be synched once and program terminates
- false: the last 30 Strava activities will be synched each 5 minutes

## Docker Container
### ... via Dockerfile
        docker build --tag stravakomootsync:latest .
        docker build --tag stravakomootsync:latest -f Dockerfile.multistage .

        docker run -d -p 8080:8080 --name stravakomootsync --restart unless-stopped --rm -e 'KOMOOT_EMAIL=*****' -e 'KOMOOT_PWD=*****' -e 'KOMOOT_USERID=*****' -e 'STRAVA_CLIENTID=*****' -e 'STRAVA_CLIENTSECRET=*****' -e 'STRAVA_ATHLETEID=*****' stravakomootsync

### ... via docker-compose
        docker-compose up -d --build

## ToDos
- sync pics
