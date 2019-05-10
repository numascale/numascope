package main

import (
//   "reflect"
   "fmt"
   "time"
   "net/http"
   "github.com/gorilla/websocket"
)

var (
   upgrader = websocket.Upgrader{}
   connections []*websocket.Conn
)

type SignonMessage struct {
   Interval int
   Tree     []map[string][]string
}

type ChangeMessage struct {
   Op        string
   Timestamp uint64
   Enabled   []string
}

type UpdateMessage struct {
   Op        string
   Timestamp uint64
   Values    []int64
}

func change(c *websocket.Conn) {
   msg := ChangeMessage{
      Op: "enabled",
      Timestamp: uint64(time.Now().UnixNano() / 1e6),
   }

   for _, sensor := range present {
      for _, event := range sensor.Events() {
         if event.enabled {
            msg.Enabled = append(msg.Enabled, event.desc)
         }
      }
   }

   c.WriteJSON(&msg)
}

func update(timestamp uint64, samples []int64) {
   msg := UpdateMessage{
      Op: "update",
      Timestamp: timestamp,
      Values: samples,
   }

   for _, c := range connections {
      err := c.WriteJSON(&msg)
      if err != nil && *debug {
         fmt.Println("failed writing: ", err)
      }
   }
}

func remove(c *websocket.Conn) {
   for i := range connections {
      if connections[i] == c {
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

      for i := range events {
         if events[i].desc == desc {
            events[i].enabled = state
            sensor.Enable(*discrete)
            return
         }
      }
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
      change(c)
   }
}

func monitor(w http.ResponseWriter, r *http.Request) {
   c, err := upgrader.Upgrade(w, r, nil)
   if err != nil {
      if *debug {
         fmt.Print("upgrade:", err)
      }
      return
   }

   defer c.Close()

   // handshake
   _, message, err := c.ReadMessage()
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
      Interval: interval,
      Tree: make([]map[string][]string, len(present)),
   }

   for i, sensor := range present {
      msg.Tree[i] = make(map[string][]string)
      name := sensor.Name()
      events := sensor.Events()

      msg.Tree[i][name] = make([]string, len(events))

      for j, val := range events {
         msg.Tree[i][name][j] = val.desc
      }
   }

   err = c.WriteJSON(&msg)
   if err != nil {
      if *debug {
         fmt.Println("write:", err)
      }
      return
   }

   change(c);
   connections = append(connections, c)

   for {
      var msg map[string]string
      err := c.ReadJSON(&msg)

      if err != nil {
         if *debug {
            fmt.Println("failed reading:", err)
         }
         remove(c)
         break
      }

      if *debug {
         fmt.Printf("recv %#v\n", msg)
      }

      switch msg["Op"] {
      case "update":
         toggle(msg["Event"], msg["State"])
      default:
         fmt.Printf("received unknown message %+v", msg)
      }
   }
}

func initweb(addr string) {
   fileServer := http.FileServer(http.Dir("resources"))
   http.Handle("/", fileServer)
   http.HandleFunc("/monitor", monitor)

   go http.ListenAndServe(addr, nil)
   fmt.Printf("interface available at http://%v\n", addr)
}
