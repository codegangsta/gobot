package gobot

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/martini-contrib/cors"
)

type startFuncAlias func(*api)

type Command func(params map[string]interface{}) []interface{}

// Optional restful API through the master to access
// all the robots.
type api struct {
	master    *Master
	server    *martini.ClassicMartini
	Host      string
	Port      string
	Username  string
	Password  string
	Cert      string
	Key       string
	startFunc startFuncAlias
}

func NewApi() *api {
	return &api{startFunc: defaultStartFunc}
}

var defaultStartFunc = func(a *api) {
	if a == nil {
		return
	}

	username := a.Username
	if username != "" {
		password := a.Password
		a.server.Use(auth.Basic(username, password))
	}

	port := a.Port
	if port == "" {
		port = "3000"
	}

	host := a.Host
	cert := a.Cert
	key := a.Key

	log.Println("Initializing API on " + host + ":" + port + "...")
	go func() {
		if cert != "" && key != "" {
			http.ListenAndServeTLS(host+":"+port, cert, key, a.server)
		} else {
			log.Println("WARNING: API using insecure connection. We recommend using an SSL certificate with Gobot.")
			http.ListenAndServe(host+":"+port, a.server)
		}
	}()
}

// start starts the api using the start function
// sets on the API on initialization.
func (a *api) start() {
	if a == nil {
		return
	}
	a.startFunc(a)
}

func Api(bot *Master) *api {
	a := new(api)
	a.master = bot
	bot.Api = a

	m := martini.Classic()
	a.server = m

	m.Use(martini.Static("robeaux"))
	m.Use(cors.Allow(&cors.Options{
		AllowAllOrigins: true,
	}))

	m.Get("/robots", func(res http.ResponseWriter, req *http.Request) {
		a.robots(res, req)
	})

	m.Get("/robots/:robotname", func(params martini.Params, res http.ResponseWriter, req *http.Request) {
		a.robot(params["robotname"], res, req)
	})

	m.Get("/robots/:robotname/commands", func(params martini.Params, res http.ResponseWriter, req *http.Request) {
		a.robot_commands(params["robotname"], res, req)
	})

	robot_command_route := "/robots/:robotname/commands/:command"

	m.Get(robot_command_route, func(params martini.Params, res http.ResponseWriter, req *http.Request) {
    bot := a.master.FindRobot(params["robotname"])
    if c := bot.GetCommand(params["command"]]); c != nil {
      c(res, req)
    } else {
      http.Error(res, "Not Found", 404)
    }
		// a.executeRobotCommand(params["robotname"], params["command"], res, req)
	})
	m.Post(robot_command_route, func(params martini.Params, res http.ResponseWriter, req *http.Request) {
		a.executeRobotCommand(params["robotname"], params["command"], res, req)
	})

	m.Get("/robots/:robotname/devices", func(params martini.Params, res http.ResponseWriter, req *http.Request) {
		a.robot_devices(params["robotname"], res, req)
	})

	m.Get("/robots/:robotname/devices/:devicename", func(params martini.Params, res http.ResponseWriter, req *http.Request) {
		a.robot_device(params["robotname"], params["devicename"], res, req)
	})

	m.Get("/robots/:robotname/devices/:devicename/commands", func(params martini.Params, res http.ResponseWriter, req *http.Request) {
		a.robot_device_commands(params["robotname"], params["devicename"], res, req)
	})

	command_route := "/robots/:robotname/devices/:devicename/commands/:command"

	m.Get(command_route, func(params martini.Params, res http.ResponseWriter, req *http.Request) {
		a.executeCommand(params["robotname"], params["devicename"], params["command"], res, req)
	})
	m.Post(command_route, func(params martini.Params, res http.ResponseWriter, req *http.Request) {
		a.executeCommand(params["robotname"], params["devicename"], params["command"], res, req)
	})

	m.Get("/robots/:robotname/connections", func(params martini.Params, res http.ResponseWriter, req *http.Request) {
		a.robot_connections(params["robotname"], res, req)
	})

	m.Get("/robots/:robotname/connections/:connectionname", func(params martini.Params, res http.ResponseWriter, req *http.Request) {
		a.robot_connection(params["robotname"], params["connectionname"], res, req)
	})

	return a
}

func (me *api) robots(res http.ResponseWriter, req *http.Request) {
	jsonRobots := make([]*jsonRobot, 0)
	for _, robot := range me.master.Robots {
		jsonRobots = append(jsonRobots, me.formatJsonRobot(robot))
	}
	data, _ := json.Marshal(jsonRobots)
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Write(data)
}

func (me *api) robot(name string, res http.ResponseWriter, req *http.Request) {
	data, _ := json.Marshal(me.formatJsonRobot(me.master.FindRobot(name)))
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Write(data)
}

func (me *api) robot_commands(name string, res http.ResponseWriter, req *http.Request) {
	data, _ := json.Marshal(me.master.FindRobot(name).RobotCommands)
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Write(data)
}

func (me *api) robot_devices(name string, res http.ResponseWriter, req *http.Request) {
	devices := me.master.FindRobot(name).GetDevices()
	jsonDevices := make([]*jsonDevice, 0)
	for _, device := range devices {
		jsonDevices = append(jsonDevices, me.formatJsonDevice(device))
	}
	data, _ := json.Marshal(jsonDevices)
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Write(data)
}

func (me *api) robot_device(robot string, device string, res http.ResponseWriter, req *http.Request) {
	data, _ := json.Marshal(me.formatJsonDevice(me.master.FindRobot(robot).GetDevice(device)))
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Write(data)
}

func (me *api) robot_device_commands(robot string, device string, res http.ResponseWriter, req *http.Request) {
	data, _ := json.Marshal(me.master.FindRobot(robot).GetDevice(device).Commands())
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Write(data)
}

func (me *api) robot_connections(name string, res http.ResponseWriter, req *http.Request) {
	connections := me.master.FindRobot(name).GetConnections()
	jsonConnections := make([]*jsonConnection, 0)
	for _, connection := range connections {
		jsonConnections = append(jsonConnections, me.formatJsonConnection(connection))
	}
	data, _ := json.Marshal(jsonConnections)
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Write(data)
}

func (me *api) robot_connection(robot string, connection string, res http.ResponseWriter, req *http.Request) {
	data, _ := json.Marshal(me.formatJsonConnection(me.master.FindRobot(robot).GetConnection(connection)))
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Write(data)
}

func (a *api) formatJsonRobot(robot *Robot) *jsonRobot {
	jsonRobot := new(jsonRobot)
	jsonRobot.Name = robot.Name
	jsonRobot.Commands = robot.RobotCommands
	jsonRobot.Connections = make([]*jsonConnection, 0)
	for _, device := range robot.devices {
		jsonDevice := a.formatJsonDevice(device)
		jsonRobot.Connections = append(jsonRobot.Connections, jsonDevice.Connection)
		jsonRobot.Devices = append(jsonRobot.Devices, jsonDevice)
	}
	return jsonRobot
}

func (a *api) formatJsonConnection(connection *connection) *jsonConnection {
	jsonConnection := new(jsonConnection)
	jsonConnection.Name = connection.Name
	jsonConnection.Port = connection.Port
	jsonConnection.Adaptor = connection.Type
	return jsonConnection
}

func (a *api) formatJsonDevice(device *device) *jsonDevice {
	jsonDevice := new(jsonDevice)
	jsonDevice.Name = device.Name
	jsonDevice.Driver = device.Type
	jsonDevice.Connection = a.formatJsonConnection(
		a.master.FindRobot(device.Robot.Name).
			GetConnection(FieldByNamePtr(FieldByNamePtr(device.Driver, "Adaptor").
			Interface().(AdaptorInterface), "Name").
			Interface().(string)))
	jsonDevice.Commands = FieldByNamePtr(device.Driver, "Commands").Interface().([]string)
	return jsonDevice
}

func (a *api) executeCommand(robotname string, devicename string, commandname string, res http.ResponseWriter, req *http.Request) {
	data, _ := ioutil.ReadAll(req.Body)
	var body map[string]interface{}
	json.Unmarshal(data, &body)
	robot := a.master.FindRobot(robotname).GetDevice(devicename)
	commands := robot.Commands().([]string)
	for command := range commands {
		if commands[command] == commandname {
			ret := make([]interface{}, 0)
			for _, v := range Call(robot.Driver, commandname, body) {
				ret = append(ret, v.Interface())
			}
			data, _ = json.Marshal(ret)
			res.Header().Set("Content-Type", "application/json; charset=utf-8")
			res.Write(data)
			return
		}
	}
	data, _ = json.Marshal([]interface{}{"Unknown Command"})
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Write(data)
}

func (a *api) executeRobotCommand(robotname string, commandname string, res http.ResponseWriter, req *http.Request) {
	data, _ := ioutil.ReadAll(req.Body)
	body := make(map[string]interface{})
	json.Unmarshal(data, &body)
	robot := a.master.FindRobot(robotname)

	in := make([]reflect.Value, 1)
	body["robotname"] = robotname
	in[0] = reflect.ValueOf(body)

	command := robot.Commands[commandname]
	if command != nil {
		data, _ = json.Marshal(command(body))
	} else {
		data, _ = json.Marshal([]interface{}{"Unknown Command"})
	}
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	res.Write(data)
}
