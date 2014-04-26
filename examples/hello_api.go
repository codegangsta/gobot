package main

import (
	"fmt"
	"github.com/hybridgroup/gobot"
)

func Hello(params map[string]interface{}) []interface{} {
	name := params["name"].(string)
	return fmt.Sprintf("hi %v", name)
}

func main() {
	master := gobot.GobotMaster()
	api := gobot.Api(master)
	api.Port = "4000"

	bot := new(gobot.Robot)
	bot.Name = "hellobot"

  hello.AddCommand("/v1/hello", func(rw http.ResponseWriter, r *http.Request) {
    friend := master.FindRobot(r.URL.Values)

    // do robot stuff
    bot.FistBump(friend)

    // serialize response
    rw.Write("robot stuff")
  })

	master.Robots = append(master.Robots, hello)

	master.Start()
}

