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
				log.Print(page.Error)
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
	if len(idsList) == 0 {
		log.Printf("No users to check, provide another input")
		return
	}

	var blocked []anaconda.User

	pages := int(math.Ceil(float64(len(idsList) / 100)))
	log.Printf("Checking %d users, %d page(s) in total", len(idsList), pages+1)

	for i := 0; i <= pages; i++ {
		log.Printf("Working on page %d/%d", i+1, pages+1)

		max := (i + 1) * 100
		if len(idsList) < max {
			max = len(idsList)
		}

		u, err := tapi.GetUsersLookupByIds(idsList[i*100:max], url.Values{})
		if err != nil {
			log.Print(err)
			continue
		}

		for _, us := range u {
			if us.Status == nil && us.StatusesCount > 0 && !us.Following && us.Id != me.Id {
				if us.Protected || us.StatusesCount < 100 {
					v := url.Values{}
					v.Add("source_id", me.IdStr)
					v.Add("target_id", us.IdStr)
					r, err := tapi.GetFriendshipsShow(v)
					if err != nil {
						log.Print(err)
					}
					if r.Relationship.Source.Blocked_By {
						blocked = append(blocked, us)
					}
				} else {
					blocked = append(blocked, us)
				}
			}
		}
	}

	for _, user := range blocked {
		log.Printf("Blocked by @%s (%s)", user.ScreenName, user.Name)
	}
}
