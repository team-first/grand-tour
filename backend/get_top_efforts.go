//   You can find an access_token for your app at https://www.strava.com/settings/api
// Grab the top efforts for each effort across each segment and sum their best
package main

import (
    "flag"
    "fmt"
    "os"
	"sort"

    "github.com/strava/go.strava"
)

type EffortsMap map[int64] [] int

type AthleteSum struct {
	athlete_id int64
	sum_time int
}

type ByEffort []AthleteSum

func (a ByEffort) Len() int           { return len(a) }
func (a ByEffort) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByEffort) Less(i, j int) bool { return a[i].sum_time < a[j].sum_time }

// now we have a map of athlete -> best efforts. Lets sum them all and find the fastest.
func sortAndPrintEffortsMap (effortsMap EffortsMap) {
	v := make([]AthleteSum, 0, len(effortsMap))
	for athlete,efforts := range effortsMap {
		// if they have a zero effort then ignore them

		sum := 0

		for _, value := range efforts {
			if value == 0 {
				sum = -1
				break
			}
			sum += value
		}

		if sum != -1 {
			fmt.Println("%d made it!", athlete)
			v = append(v,AthleteSum{athlete, sum})
		}
	}

	sort.Sort(ByEffort(v))
	fmt.Println("Top 10 efforts: ")
	x:=10
	if x > len(v) {
		x = len(v)
	}

	for i := range v[0:x] {
		fmt.Println("%d:%d",v[i].athlete_id, v[i].sum_time)
	}
}




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
    var effortsMap = make(EffortsMap)
	// TODO: make this configurable via a tour Id command line param
    var segments = [] int64 {
        2995991,  // ek
		3377798, //wb
        6047900, // man
        640467 } // bk

    for segmentIdx, segmentId := range segments {
        fmt.Printf("Fetching segment %d info...\n", segmentId)

        for i := 1;i<2;i++ {
			// only look for 30 days and get 200 people.
			results, err := strava.NewSegmentsService(client).
				GetLeaderboard(int64(segmentId)).
				Page(i).
				PerPage(200).
				DateRange("this_month").
				Do()

			if err != nil {
                fmt.Println(err)
                fmt.Printf("Break\n")
                break
			 }

            for _, e := range results.Entries {
                fmt.Printf("%s: %5d %5d\n", e.AthleteName, e.MovingTime, e.AthleteId)

                cur, ok := effortsMap[e.AthleteId]
				// If this is the first time we see this athlete then create them a new array
                if !ok {
                    a := make([]int,len(segments))
                    a[segmentIdx] = e.MovingTime
                    effortsMap[e.AthleteId] = a
                } else if cur[segmentIdx] > e.MovingTime {
                    fmt.Printf("beat! %5d %5d\n", cur[segmentIdx], e.MovingTime)
                    cur[segmentIdx] = e.MovingTime
                    effortsMap[e.AthleteId] = cur
                }
            }
        }
    }

	sortAndPrintEffortsMap(effortsMap)
}
