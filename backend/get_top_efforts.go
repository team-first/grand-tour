//   You can find an access_token for your app at https://www.strava.com/settings/api
// Grab the top efforts for each effort across each segment and sum their best
package main

import (
    "flag"
    "fmt"
    "os"
    "time"

    "github.com/strava/go.strava"
)


func main() {
    var accessToken string

    // Provide an access token, with write permissions.
    // You'll need to complete the oauth flow to get one.
    flag.StringVar(&accessToken, "token", "", "Access Token")

    flag.Parse()

    if accessToken == "" {
        fmt.Println("\nPlease provide an access_token, one can be found at https://www.strava.com/settings/api")

        flag.PrintDefaults()
        os.Exit(1)
    }

    client := strava.NewClient(accessToken)
    // make a map per athlete to their list of efforts
    var effortsMap = make(map[int64] [4] int)
    // var segments = []int64 {
    //     6401036,
    //     2622770,
    //     6400995,
    //     2276683 }
    var segments = [] int64 {
        2995991,
        1112424,
        6047900,
        640467    }
    for segmentIdx, segmentId := range segments {
        fmt.Printf("Fetching segment %d info...\n", segmentId)

        for i := 1;;i++ {

            results, err := strava.NewSegmentsService(client).ListEfforts(int64(segmentId)).DateRange(time.Now().Add(-30*24*time.Hour),time.Now()).PerPage(100).Page(i).Do()
             if err != nil {
                fmt.Println(err)
                fmt.Printf("Break\n")
                break
               }

            for _, e := range results {
                fmt.Printf("%5d: %5d %5d\n", e.Id, e.MovingTime, e.Athlete.Id)

                cur, ok := effortsMap[e.Athlete.Id]

                if !ok {
                    var a [4] int
                    a[segmentIdx] = e.MovingTime
                    effortsMap[e.Athlete.Id] = a
                } else if cur[segmentIdx] > e.MovingTime {
                    fmt.Printf("beat! %5d %5d\n", cur[segmentIdx], e.MovingTime)
                    cur[segmentIdx] = e.MovingTime
                    effortsMap[e.Athlete.Id] = cur
                }
            }
        }
    }
}
