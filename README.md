# StravaKomootSync

[![Docker Image CI](https://github.com/aexel90/strava_komoot_sync/actions/workflows/docker-image.yml/badge.svg)](https://github.com/aexel90/strava_komoot_sync/actions/workflows/docker-image.yml)

Activity Synchronization between Strava and Komoot.
Synchronization direction: Strava --> Komoot

What is synced:
- Name of the activity
- Visibility (private, public)

## Script Parameter

    $GOPATH/bin/strava_komoot_sync -h

    Usage of ./strava_komoot_sync:
        -debug
                Log debug level
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
        -strava_virtualRide_gearid string
                Strava Virtual Ride GearID
        -sync_all
                Sync all activities

Flag "-sync_all"
- true:  all activities will be synched once and program terminates
- false: the last 30 Strava activities will be synched

## Run Docker Container
### ... via Dockerfile
        docker build --tag stravakomootsync:latest .
        docker run -d -p 8080:8080 --name stravakomootsync --restart unless-stopped -e 'KOMOOT_EMAIL=*****' -e 'KOMOOT_PWD=*****' -e 'KOMOOT_USERID=*****' -e 'STRAVA_CLIENTID=*****' -e 'STRAVA_CLIENTSECRET=*****' -e 'STRAVA_ATHLETEID=*****' -e 'STRAVA_VIRT_GEARID=*****' stravakomootsync

### ... via docker-compose and pre-build package from ghcr.io
        cp .env.template .env
        vi .env
        docker compose up -d

## TODOs
- sync pics
