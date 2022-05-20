FROM golang:1.16-alpine

WORKDIR /app

COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 8080

### SHELL FORM
#CMD "strava_komoot_sync" "-komoot_email" ${KOMOOT_EMAIL} "-komoot_pw" ${KOMOOT_PWD} "-komoot_userid" ${KOMOOT_USERID} "-strava_clientid" ${STRAVA_CLIENTID} "-strava_clientsecret" ${STRAVA_CLIENTSECRET} "-strava_athleteid" ${STRAVA_ATHLETEID}

### EXEC FORM
ENTRYPOINT [ "sh", "-c", "strava_komoot_sync -komoot_email ${KOMOOT_EMAIL} -komoot_pw ${KOMOOT_PWD} -komoot_userid ${KOMOOT_USERID} -strava_clientid ${STRAVA_CLIENTID} -strava_clientsecret ${STRAVA_CLIENTSECRET} -strava_athleteid ${STRAVA_ATHLETEID}" ]