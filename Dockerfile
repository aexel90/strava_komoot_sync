FROM golang:alpine AS build
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . .
RUN go build -o /strava_komoot_sync

FROM alpine:latest
WORKDIR /
COPY --from=build /strava_komoot_sync /strava_komoot_sync
EXPOSE 8080
ENTRYPOINT [ "sh", "-c", "/strava_komoot_sync -komoot_email ${KOMOOT_EMAIL} -komoot_pw ${KOMOOT_PWD} -komoot_userid ${KOMOOT_USERID} -strava_clientid ${STRAVA_CLIENTID} -strava_clientsecret ${STRAVA_CLIENTSECRET} -strava_athleteid ${STRAVA_ATHLETEID} -strava_virtualRide_gearid ${STRAVA_VIRT_GEARID}" ]