package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/Sirupsen/logrus"
)

const (
	APP     = "twitter-cleaner v%s\n"
	VERSION = "1.0.0"
)

var (
	consumerKey       string
	consumerSecret    string
	accessToken       string
	accessTokenSecret string
	age               int
	debug             bool
	now               = time.Now()
	api               *anaconda.TwitterApi
)

func init() {
	flag.StringVar(&consumerKey, "consumerKey", "", "Your consumer key")
	flag.StringVar(&consumerSecret, "consumerSecret", "", "Your consumer secret")
	flag.StringVar(&accessToken, "accessToken", "", "Your access token")
	flag.StringVar(&accessTokenSecret, "accessTokenSecret", "", "Your access token secret")
	flag.IntVar(&age, "age", 61, "How many days since last tweet ?")
	flag.BoolVar(&debug, "d", false, "Run in debug mode")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(APP, VERSION))
		flag.PrintDefaults()
	}

	flag.Parse()

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func main() {
	logrus.Debugf("consumerKey %s", consumerKey)
	logrus.Debugf("consumerSecret %s", consumerSecret)
	logrus.Debugf("accessToken %s", accessToken)
	logrus.Debugf("accessTokenSecret %s", accessTokenSecret)
	logrus.Debugf("age %d", age)
	logrus.Debugf("debug %t", debug)

	anaconda.SetConsumerKey(consumerKey)
	anaconda.SetConsumerSecret(consumerSecret)
	api = anaconda.NewTwitterApi(accessToken, accessTokenSecret)

	cursor, err := api.GetFriendsIds(nil)
	if err != nil {
		logrus.Fatal(err)
		return
	}

	for _, id := range cursor.Ids {
		processId(id)
	}
}

func processId(id int64) {
	var idArray []int64
	idArray = append(idArray, id)
	user, err := api.GetUsersLookupByIds(idArray, nil)
	if err != nil {
		logrus.Warnf("Failed to read user %s", id)
		logrus.Debug(err)
		return
	}

	getUserTimeline(user[0])
}

func getUserTimeline(user anaconda.User) {
	logrus.Debugf("Checking %s (%s)", user.Name, user.ScreenName)

	v := url.Values{}
	v.Set("screen_name", user.ScreenName)

	timeline, err := api.GetUserTimeline(v)
	if err != nil {
		logrus.Warnf("Failed to get timeline for %s", user.ScreenName)
		logrus.Debug(err)
		return
	}

	if len(timeline) > 0 {
		shouldRemoveQuestionMark(timeline, user)
	}
}

func shouldRemoveQuestionMark(timeline []anaconda.Tweet, user anaconda.User) {
	tweet := timeline[0]

	tweetTime, err := tweet.CreatedAtTime()
	if err != nil {
		logrus.Warnf("Failed to get tweet time")
		logrus.Debug(err)
		return
	}

	daysFromLastTweet := int(now.Sub(tweetTime).Hours() / 24)
	logrus.Debugf("Days from last tweet %d", daysFromLastTweet)

	if daysFromLastTweet > age {
		fmt.Printf("Unfollow %s (last seen: %s) ? (y/N) ", user.ScreenName, tweetTime)

		var toDelete string
		if fmt.Scanf("%s", &toDelete); toDelete == "y" {
			doRemove(user)
		}
	}
}

func doRemove(user anaconda.User) {
	logrus.Debugf("Goodbye old friend! %s (%s)", user.Name, user.ScreenName)
	_, err := api.UnfollowUserId(user.Id)
	if err != nil {
		fmt.Println("Failed to unfollow !", err)
	}
}
