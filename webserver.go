package main

import (
   "fmt"
   "net/http"
   "strconv"
   "strings"
   "sync"
   "time"

   "github.com/gorilla/websocket"
   "golang.org/x/sys/unix"
)

type SignonMessage struct {
   Tree     map[string][]string
   Sources  map[string]uint
}

type ChangeMessage struct {
   Op        string
   Timestamp uint64
   Interval int
   Discrete  bool
   Enabled   map[string][]string
}

type DataMessage struct {
   Op        string
   Timestamp uint64
   Values    []int64
}

type LabelMessage struct {
   Op        string
   Timestamp uint64
   Label     string
}

type Connection struct {
   socket  *websocket.Conn
   mutex   *sync.Mutex
   stopped bool
}

var (
   upgrader = websocket.Upgrader{}
   connections []*Connection
)

func (c *Connection) WriteJSON(msg interface{}) error {
   if *debug {
      fmt.Printf("-> %+v\n", msg)
   }

   c.mutex.Lock()
   err := c.socket.WriteJSON(msg)
   c.mutex.Unlock()

   return err
}

func change(c Connection) {
   msg := ChangeMessage{
      Op: "enabled",
      Timestamp: uint64(time.Now().UnixNano() / 1e6),
      Interval: interval,
      Discrete: *discrete,
      Enabled: make(map[string][]string),
   }

   // structure events into hashmap
   for _, sensor := range present {
      name := sensor.Name()
      msg.Enabled[name] = make([]string, 0, 16)

      for _, event := range sensor.Events() {
         if event.enabled {
            msg.Enabled[name] = append(msg.Enabled[name], event.desc)
         }
      }
   }

   err := c.WriteJSON(&msg)
   if err != nil && *debug {
      fmt.Println("failed writing:", err)
   }
}

func broadcastLabel(timestamp uint64, label string) {
   msg := LabelMessage{
      Op: "label",
      Timestamp: timestamp,
      Label: label,
   }

   for _, c := range connections {
      err := c.WriteJSON(&msg)
      if err != nil && *debug {
         fmt.Println("failed writing:", err)
      }
   }
}

func broadcastData(timestamp uint64, samples []int64) {
   msg := DataMessage{
      Op: "data",
      Timestamp: timestamp,
      Values: samples,
   }

   for _, c := range connections {
      if c.stopped {
         continue
      }

      err := c.WriteJSON(&msg)

      if err != nil && *debug {
         fmt.Println("failed writing:", err)
      }
   }
}

func remove(c *websocket.Conn) {
   for i := range connections {
      if connections[i].socket == c {
         connections[i] = connections[len(connections)-1]
         connections = connections[:len(connections)-1]
         return
      }
   }

   panic("element not found")
}

func state(desc string, state bool) {
   for _, sensor := range present {
      events := sensor.Events()
      sensor.Lock()

      // check if 'all' button was selected
      if desc == /*sensor.Name() +*/ "all" {
         for i := range events {
            events[i].enabled = true
         }

         sensor.Enable(*discrete)
         sensor.Unlock()
         // discard values to initialise last
         sensor.Sample()
         return
      }

      for i := range events {
         if events[i].desc == desc {
            events[i].enabled = state
            sensor.Enable(*discrete)
            sensor.Unlock()
            // discard values to initialise last
            sensor.Sample()
            return
         }
      }

      sensor.Unlock()
   }

   panic("event not found")
}

func toggle(desc string, val string) {
   switch (val) {
   case "on":
      state(desc, true)
   case "off":
      state(desc, false)
   default:
      panic("unexpected state")
   }

   // update all clients
   for _, c := range connections {
      change(*c)
   }
}

func monitor(w http.ResponseWriter, r *http.Request) {
   socket, err := upgrader.Upgrade(w, r, nil)
   if err != nil {
      if *debug {
         fmt.Print("upgrade:", err)
      }
      return
   }

   defer socket.Close()

   c := Connection{socket: socket, mutex: &sync.Mutex{}}

   // handshake
   _, message, err := c.socket.ReadMessage()
   if err != nil {
      if *debug {
         fmt.Println("read:", err)
      }
      return
   }

   if string(message) != "463ba1974b06" {
      if *debug {
         fmt.Println("auth failed")
      }
      return
   }

   if *debug {
      fmt.Println("auth succeeded")
   }

   msg := SignonMessage{
      Tree: make(map[string][]string, len(present)),
      Sources: make(map[string]uint, len(present)),
   }

   msg.Tree = make(map[string][]string)

   for _, sensor := range present {
      name := sensor.Name()
      events := sensor.Events()

      msg.Tree[name] = make([]string, len(events))
      msg.Sources[name] = sensor.Sources()

      for i, val := range events {
         msg.Tree[name][i] = val.desc
      }
   }

   err = c.WriteJSON(&msg)
   if err != nil {
      if *debug {
         fmt.Println("failed writing:", err)
      }
      return
   }

   change(c);
   connections = append(connections, &c)

   for {
      var msg map[string]string
      err := c.socket.ReadJSON(&msg)

      if err != nil {
         if *debug {
            fmt.Println("failed reading:", err)
         }
         remove(c.socket)
         break
      }

      if *debug {
         fmt.Printf("recv %#v\n", msg)
      }

      switch msg["Op"] {
      case "update":
         toggle(msg["Event"], msg["State"])
      case "stop":
         c.stopped = true
      case "start":
         c.stopped = false
      case "averaging":
         *discrete = msg["Value"] == "false"
         Activate()

         for _, c2 := range connections {
            change(*c2)
         }
      case "interval":
         interval, err = strconv.Atoi(msg["Value"])
         if err != nil {
            fmt.Printf("undefined value %v\n", msg["Value"])
         }
      default:
         fmt.Printf("received unknown message %+v\n", msg)
      }
   }
}

func initweb(addr string) {
   path := "/usr/local/share/numascope"
   err := unix.Access(path, unix.R_OK)
   if err != nil {
      path = "resources"
      err := unix.Access(path, unix.R_OK)
      if err != nil {
         panic("/usr/local/share/numascope or resources not present")
      }
   }

   fileServer := http.FileServer(http.Dir(path))
   http.Handle("/", fileServer)
   http.HandleFunc("/monitor", monitor)

   go http.ListenAndServe(addr, nil)
   port := strings.Split(addr, ":")[1]
   fmt.Printf("web interface available on port %s\n", port)
}
