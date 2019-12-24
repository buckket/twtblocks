package main

import (
	"flag"
	"github.com/buckket/anaconda"
	"github.com/spf13/viper"
	"log"
	"math"
	"net/url"
)

func main() {
	configPtr := flag.String("config", "", "path to config file")
	flag.Parse()

	if len(*configPtr) > 0 {
		viper.SetConfigFile(*configPtr)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
	}

	err := viper.ReadInConfig()
	if err != nil {
		log.Print(err)
	}
	viper.AutomaticEnv()

	tapi := anaconda.NewTwitterApiWithCredentials(viper.GetString("TWITTER_ACCESS_TOKEN"),
		viper.GetString("TWITTER_ACCESS_TOKEN_SECRET"),
		viper.GetString("TWITTER_CONSUMER_KEY"),
		viper.GetString("TWITTER_CONSUMER_SECRET"))
	me, err := tapi.GetSelf(url.Values{})
	if err != nil {
		log.Fatal(err)
	}

	userlist := flag.Args()
	if len(userlist) == 0 {
		log.Fatal("Please provide a list of users as a starting point")
	}

	idsSet := make(map[int64]bool)
	for _, user := range userlist {
		v := url.Values{}
		v.Add("screen_name", user)
		c := tapi.GetFriendsIdsAll(v)
		for page := range c {
			if page.Error != nil {
				log.Fatal(err)
			}
			for _, id := range page.Ids {
				idsSet[id] = true
			}
		}
	}

	idsList := make([]int64, 0, len(idsSet))
	for k := range idsSet {
		idsList = append(idsList, k)
	}

	pages := int(math.Ceil(float64(len(idsList) / 100)))
	for i := 0; i <= pages; i++ {
		max := (i+1)*100
		if len(idsList) < max {
			max = len(idsList)
		}
		u, err := tapi.GetUsersLookupByIds(idsList[i*100:max], url.Values{})
		if err != nil {
			log.Fatal(err)
		}
		for _, us := range u {
			if us.Status == nil && us.StatusesCount > 0 && !us.Following && us.Id != me.Id {
				if us.Protected || us.StatusesCount < 100 {
					v := url.Values{}
					v.Add("source_id", me.IdStr)
					v.Add("target_id", us.IdStr)
					r, err := tapi.GetFriendshipsShow(v)
					if err != nil {
						log.Fatal(err)
					}
					if r.Relationship.Source.Blocked_By {
						log.Printf("Blocked by @%s (%s)", us.ScreenName, us.Name)
					}
				} else {
					log.Printf("Blocked by @%s (%s)", us.ScreenName, us.Name)
				}
			}
		}
	}
}
